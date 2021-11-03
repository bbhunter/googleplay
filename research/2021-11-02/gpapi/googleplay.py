from . import googleplay_pb2
from base64 import b64decode, urlsafe_b64encode
from cryptography.hazmat.backends import default_backend
from cryptography.hazmat.primitives import hashes
from cryptography.hazmat.primitives.asymmetric import padding
from cryptography.hazmat.primitives.asymmetric.utils import encode_dss_signature
from cryptography.hazmat.primitives.serialization import load_der_public_key
from datetime import datetime
from os import path
from re import match
from sys import version_info
from time import time
import requests

BASE = "https://android.clients.google.com/"
FDFE = BASE + "fdfe/"

AUTH_URL = BASE + "auth"
BROWSE_URL = FDFE + "browse"
BULK_URL = FDFE + "bulkDetails"
CHECKIN_URL = BASE + "checkin"
CONTENT_TYPE_PROTO = "application/x-protobuf"
CONTENT_TYPE_URLENC = "application/x-www-form-urlencoded; charset=UTF-8"
DELIVERY_URL = FDFE + "delivery"
DETAILS_URL = FDFE + "details"
HOME_URL = FDFE + "homeV2"
LIST_URL = FDFE + "list"
PURCHASE_URL = FDFE + "purchase"
REVIEWS_URL = FDFE + "rev"
SEARCH_URL = FDFE + "search"
UPLOAD_URL = FDFE + "uploadDeviceConfig"
ssl_verify = True

class LoginError(Exception):
    def __init__(self, value):
        self.value = value

    def __str__(self):
        return repr(self.value)

class GooglePlayAPI(object):
    """Google Play Unofficial API Class
    Usual APIs methods are login(), search(), details(), bulkDetails(),
    download(), browse(), reviews() and list()."""

    def __init__(self, locale="en_US", timezone="UTC", device_codename="bacon",
                 proxies_config=None):
        self.authSubToken = None
        self.gsfId = None
        self.device_config_token = None
        self.deviceCheckinConsistencyToken = None
        self.dfeCookie = None
        self.proxies_config = proxies_config
        self.deviceBuilder = DeviceBuilder(device_codename)
        self.setLocale(locale)
        self.setTimezone(timezone)

    def setLocale(self, locale):
        self.deviceBuilder.setLocale(locale)

    def setTimezone(self, timezone):
        self.deviceBuilder.setTimezone(timezone)

    def setAuthSubToken(self, authSubToken):
        self.authSubToken = authSubToken

    def getHeaders(self, upload_fields=False):
        """Return the default set of request headers, which
        can later be expanded, based on the request type"""
        if upload_fields:
            headers = self.deviceBuilder.getDeviceUploadHeaders()
        else:
            headers = self.deviceBuilder.getBaseHeaders()
        if self.gsfId is not None:
            headers["X-DFE-Device-Id"] = "{0:x}".format(self.gsfId)
        if self.authSubToken is not None:
            headers["Authorization"] = "GoogleLogin auth=%s" % self.authSubToken
        if self.device_config_token is not None:
            headers["X-DFE-Device-Config-Token"] = self.device_config_token
        if self.deviceCheckinConsistencyToken is not None:
            headers["X-DFE-Device-Checkin-Consistency-Token"] = self.deviceCheckinConsistencyToken
        if self.dfeCookie is not None:
            headers["X-DFE-Cookie"] = self.dfeCookie
        return headers

    def checkin(self, email, ac2dmToken):
        headers = self.getHeaders()
        headers["Content-Type"] = CONTENT_TYPE_PROTO
        request = self.deviceBuilder.getAndroidCheckinRequest()
        stringRequest = request.SerializeToString()
        res = requests.post(CHECKIN_URL, data=stringRequest,
                            headers=headers, verify=ssl_verify,
                            proxies=self.proxies_config)
        print(res.status_code, res.url)
        response = googleplay_pb2.AndroidCheckinResponse()
        response.ParseFromString(res.content)
        self.deviceCheckinConsistencyToken = response.deviceCheckinConsistencyToken
        # checkin again to upload gfsid
        request.id = response.androidId
        request.securityToken = response.securityToken
        request.accountCookie.append("[" + email + "]")
        request.accountCookie.append(ac2dmToken)
        stringRequest = request.SerializeToString()
        res = requests.post(CHECKIN_URL,
                      data=stringRequest,
                      headers=headers,
                      verify=ssl_verify,
                      proxies=self.proxies_config)
        print(res.status_code, res.url)
        return response.androidId

    def uploadDeviceConfig(self):
        """Upload the device configuration of the fake device
        selected in the __init__ methodi to the google account."""
        upload = googleplay_pb2.UploadDeviceConfigRequest()
        upload.deviceConfiguration.CopyFrom(self.deviceBuilder.getDeviceConfig())
        headers = self.getHeaders(upload_fields=True)
        stringRequest = upload.SerializeToString()
        response = requests.post(UPLOAD_URL, data=stringRequest,
            headers=headers,
            verify=ssl_verify,
            timeout=60,
            proxies=self.proxies_config)
        print(response.status_code, response.url)
        response = googleplay_pb2.ResponseWrapper.FromString(response.content)
        try:
            if response.payload.HasField('uploadDeviceConfigResponse'):
                self.device_config_token = response.payload.uploadDeviceConfigResponse
                self.device_config_token = self.device_config_token.uploadDeviceConfigToken
        except ValueError:
            pass

    def getAuthSubToken(self, email, passwd):
        requestParams = self.deviceBuilder.getLoginParams(email, passwd)
        requestParams['service'] = 'androidmarket'
        requestParams['app'] = 'com.android.vending'
        headers = self.deviceBuilder.getAuthHeaders(self.gsfId)
        headers['app'] = 'com.android.vending'
        response = requests.post(AUTH_URL,
            data=requestParams,
            verify=ssl_verify,
            headers=headers,
            proxies=self.proxies_config)
        print(response.status_code, response.url)
        data = response.text.split()
        params = {}
        for d in data:
            if "=" not in d:
                continue
            k, v = d.split("=", 1)
            params[k.strip().lower()] = v.strip()
        if "token" in params:
            master_token = params["token"]
            second_round_token = self.getSecondRoundToken(master_token, requestParams)
            self.setAuthSubToken(second_round_token)
        elif "error" in params:
            raise LoginError("server says: " + params["error"])
        else:
            raise LoginError("auth token not found.")

    def getSecondRoundToken(self, first_token, params):
        if self.gsfId is not None:
            params['androidId'] = "{0:x}".format(self.gsfId)
        params['Token'] = first_token
        params['check_email'] = '1'
        params['token_request_options'] = 'CAA4AQ=='
        params['system_partition'] = '1'
        params['_opt_is_called_from_account_manager'] = '1'
        params.pop('Email')
        params.pop('EncryptedPasswd')
        headers = self.deviceBuilder.getAuthHeaders(self.gsfId)
        headers['app'] = 'com.android.vending'
        response = requests.post(AUTH_URL,
            data=params,
            headers=headers,
            verify=ssl_verify,
            proxies=self.proxies_config)
        print(response.status_code, response.url)
        data = response.text.split()
        params = {}
        for d in data:
            if "=" not in d:
                continue
            k, v = d.split("=", 1)
            params[k.strip().lower()] = v.strip()
        if "auth" in params:
            return params["auth"]
        elif "error" in params:
            raise LoginError("server says: " + params["error"])
        else:
            raise LoginError("Auth token not found.")

import configparser

DFE_TARGETS = "CAEScFfqlIEG6gUYogFWrAISK1WDAg+hAZoCDgIU1gYEOIACFkLMAeQBnASLATlASUuyAyqCAjY5igOMBQzfA/IClwFbApUC4ANbtgKVAS7OAX8YswHFBhgDwAOPAmGEBt4OfKkB5weSB5AFASkiN68akgMaxAMSAQEBA9kBO7UBFE1KVwIDBGs3go6BBgEBAgMECQgJAQIEAQMEAQMBBQEBBAUEFQYCBgUEAwMBDwIBAgOrARwBEwMEAg0mrwESfTEcAQEKG4EBMxghChMBDwYGASI3hAEODEwXCVh/EREZA4sBYwEdFAgIIwkQcGQRDzQ2fTC2AjfVAQIBAYoBGRg2FhYFBwEqNzACJShzFFblAo0CFxpFNBzaAd0DHjIRI4sBJZcBPdwBCQGhAUd2A7kBLBVPngEECHl0UEUMtQETigHMAgUFCc0BBUUlTywdHDgBiAJ+vgKhAU0uAcYCAWQ/5ALUAw1UwQHUBpIBCdQDhgL4AY4CBQICjARbGFBGWzA1CAEMOQH+BRAOCAZywAIDyQZ2MgM3BxsoAgUEBwcHFia3AgcGTBwHBYwBAlcBggFxSGgIrAEEBw4QEqUCASsWadsHCgUCBQMD7QICA3tXCUw7ugJZAwGyAUwpIwM5AwkDBQMJA5sBCw8BNxBVVBwVKhebARkBAwsQEAgEAhESAgQJEBCZATMdzgEBBwG8AQQYKSMUkAEDAwY/CTs4/wEaAUt1AwEDAQUBAgIEAwYEDx1dB2wGeBFgTQ"
GOOGLE_PUBKEY = "AAAAgMom/1a/v0lblO2Ubrt60J2gcuXSljGFQXgcyZWveWLEwo6prwgi3iJIZdodyhKZQrNWp5nKJ3srRXcUW+F1BD3baEVGcmEgqaLZUNBjm057pKRI16kB0YppeGx5qIQ5QjKzsR8ETQbKLNWgRY0QRNVz34kMJR3P/LgHax/6rmf5AAAAAwEAAQ=="
ACCOUNT = "HOSTED_OR_GOOGLE"

# parse phone config from the file 'device.properties'.
# if you want to add another phone, just create another section in
# the file. Some configurations for common phones can be found here:
# https://github.com/yeriomin/play-store-api/tree/master/src/main/resources
filepath = path.join(path.dirname(path.realpath(__file__)),
                     'device.properties')

config = configparser.ConfigParser()
config.read(filepath)

class InvalidLocaleError(Exception):
    pass

class InvalidTimezoneError(Exception):
    pass

def getDevicesCodenames():
    """Returns a list containing devices codenames"""
    return config.sections()


def getDevicesReadableNames():
    """Returns codename and readable name for each device"""
    return [{'codename': s,
             'readableName': config.get(s).get('userreadablename')}
            for s in getDevicesCodenames()]


class DeviceBuilder(object):

    def __init__(self, device):
        self.device = {}
        for (key, value) in config.items(device):
            self.device[key] = value

    def setLocale(self, locale):
        # test if provided locale is valid
        if locale is None or type(locale) is not str:
            raise InvalidLocaleError()

        # check if locale matches the structure of a common
        # value like "en_US"
        if match(r'[a-z]{2}\_[A-Z]{2}', locale) is None:
            raise InvalidLocaleError()
        self.locale = locale

    def setTimezone(self, timezone):
        if timezone is None or type(timezone) is not str:
            timezone = self.device.get('timezone')
            if timezone is None:
                raise InvalidTimezoneError()
        self.timezone = timezone

    def getBaseHeaders(self):
        return {"Accept-Language": self.locale.replace('_', '-'),
                "X-DFE-Encoded-Targets": DFE_TARGETS,
                "User-Agent": self.getUserAgent(),
                "X-DFE-Client-Id": "am-android-google",
                "X-DFE-MCCMNC": self.device.get('celloperator'),
                "X-DFE-Network-Type": "4",
                "X-DFE-Content-Filters": "",
                "X-DFE-Request-Params": "timeoutMs=4000"}

    def getDeviceUploadHeaders(self):
        headers = self.getBaseHeaders()
        headers["X-DFE-Enabled-Experiments"] = "cl:billing.select_add_instrument_by_default"
        headers["X-DFE-Unsupported-Experiments"] = ("nocache:billing.use_charging_poller,"
            "market_emails,buyer_currency,prod_baseline,checkin.set_asset_paid_app_field,"
            "shekel_test,content_ratings,buyer_currency_in_app,nocache:encrypted_apk,recent_changes")
        headers["X-DFE-SmallestScreenWidthDp"] = "320"
        headers["X-DFE-Filter-Level"] = "3"
        return headers


    def getUserAgent(self):
        version_string = self.device.get('vending.versionstring')
        if version_string is None:
            version_string = '8.4.19.V-all [0] [FP] 175058788'
        return ("Android-Finsky/{versionString} ("
                "api=3"
                ",versionCode={versionCode}"
                ",sdk={sdk}"
                ",device={device}"
                ",hardware={hardware}"
                ",product={product}"
                ",platformVersionRelease={platform_v}"
                ",model={model}"
                ",buildId={build_id}"
                ",isWideScreen=0"
                ",supportedAbis={supported_abis}"
                ")").format(versionString=version_string,
                            versionCode=self.device.get('vending.version'),
                            sdk=self.device.get('build.version.sdk_int'),
                            device=self.device.get('build.device'),
                            hardware=self.device.get('build.hardware'),
                            product=self.device.get('build.product'),
                            platform_v=self.device.get('build.version.release'),
                            model=self.device.get('build.model'),
                            build_id=self.device.get('build.id'),
                            supported_abis=self.device.get('platforms').replace(',', ';'))

    def getAuthHeaders(self, gsfid):
        headers = {"User-Agent": ("GoogleAuth/1.4 ("
                                  "{device} {id}"
                                  ")").format(device=self.device.get('build.device'),
                                              id=self.device.get('build.id'))}
        if gsfid is not None:
            headers['device'] = "{0:x}".format(gsfid)
        return headers

    def getLoginParams(self, email, encrypted_passwd):
        return {"Email": email,
                "EncryptedPasswd": encrypted_passwd,
                "add_account": "1",
                "accountType": ACCOUNT,
                "google_play_services_version": self.device.get('gsf.version'),
                "has_permission": "1",
                "source": "android",
                "device_country": self.locale[0:2],
                "lang": self.locale,
                "client_sig": "38918a453d07199354f8b19af05ec6562ced5788",
                "callerSig": "38918a453d07199354f8b19af05ec6562ced5788",
               "droidguard_results": "dummy123"
               }

    def getAndroidCheckinRequest(self):
        request = googleplay_pb2.AndroidCheckinRequest()
        request.id = 0
        request.checkin.CopyFrom(self.getAndroidCheckin())
        request.locale = self.locale
        request.timeZone = self.timezone
        request.version = 3
        request.deviceConfiguration.CopyFrom(self.getDeviceConfig())
        request.fragment = 0
        return request

    def getDeviceConfig(self):
        libList = self.device['sharedlibraries'].split(",")
        featureList = self.device['features'].split(",")
        localeList = self.device['locales'].split(",")
        glList = self.device['gl.extensions'].split(",")
        platforms = self.device['platforms'].split(",")

        hasFiveWayNavigation = (self.device['hasfivewaynavigation'] == 'true')
        hasHardKeyboard = (self.device['hashardkeyboard'] == 'true')
        deviceConfig = googleplay_pb2.DeviceConfigurationProto()
        deviceConfig.touchScreen = int(self.device['touchscreen'])
        deviceConfig.keyboard = int(self.device['keyboard'])
        deviceConfig.navigation = int(self.device['navigation'])
        deviceConfig.screenLayout = int(self.device['screenlayout'])
        deviceConfig.hasHardKeyboard = hasHardKeyboard
        deviceConfig.hasFiveWayNavigation = hasFiveWayNavigation
        deviceConfig.screenDensity = int(self.device['screen.density'])
        deviceConfig.screenWidth = int(self.device['screen.width'])
        deviceConfig.screenHeight = int(self.device['screen.height'])
        deviceConfig.glEsVersion = int(self.device['gl.version'])
        for x in platforms:
            deviceConfig.nativePlatform.append(x)
        for x in libList:
            deviceConfig.systemSharedLibrary.append(x)
        for x in featureList:
            deviceConfig.systemAvailableFeature.append(x)
        for x in localeList:
            deviceConfig.systemSupportedLocale.append(x)
        for x in glList:
            deviceConfig.glExtension.append(x)
        return deviceConfig

    def getAndroidBuild(self):
        androidBuild = googleplay_pb2.AndroidBuildProto()
        androidBuild.id = self.device['build.fingerprint']
        androidBuild.product = self.device['build.hardware']
        androidBuild.carrier = self.device['build.brand']
        androidBuild.radio = self.device['build.radio']
        androidBuild.bootloader = self.device['build.bootloader']
        androidBuild.device = self.device['build.device']
        androidBuild.sdkVersion = int(self.device['build.version.sdk_int'])
        androidBuild.model = self.device['build.model']
        androidBuild.manufacturer = self.device['build.manufacturer']
        androidBuild.buildProduct = self.device['build.product']
        androidBuild.client = self.device['client']
        androidBuild.otaInstalled = False
        androidBuild.timestamp = int(time()/1000)
        androidBuild.googleServices = int(self.device['gsf.version'])
        return androidBuild

    def getAndroidCheckin(self):
        androidCheckin = googleplay_pb2.AndroidCheckinProto()
        androidCheckin.build.CopyFrom(self.getAndroidBuild())
        androidCheckin.lastCheckinMsec = 0
        androidCheckin.cellOperator = self.device['celloperator']
        androidCheckin.simOperator = self.device['simoperator']
        androidCheckin.roaming = self.device['roaming']
        androidCheckin.userNumber = 0
        return androidCheckin
