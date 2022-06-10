# Docs

## APK to Java

~~~
jadx.bat com.google.android.youtube-1528288704.apk
~~~

https://github.com/skylot/jadx

## App category

https://github.com/89z/googleplay/tree/4ee083e441c3d183c9c1db9da006849630305ba6

## Geo-blocking

Some apps are specific to region. For example, `air.ITVMobilePlayer` is specifc
to GB. If you try it from US, details will work, but delivery will fail:

~~~
> googleplay -a air.ITVMobilePlayer
Title: ITV Hub: Your TV Player - Watch Live & On Demand
UploadDate: Dec 9, 2021
VersionString: 9.19.0
VersionCode: 901900000
NumDownloads: 17.429 M
Size: 35.625 MB
Offer: 0.00 USD

> googleplay -a air.ITVMobilePlayer -v 901900000
panic: Geo-blocking
~~~

It seems headers are ignored as well:

~~~
Accept-Language: es
Accept-Language: es-AR
Accept-Language: es-ES
~~~

You can change the country [1], and then you get expected result:

~~~
> googleplay -a air.ITVMobilePlayer
Title: ITV Hub: Your TV Player - Watch Live & On Demand
UploadDate: Dec 9, 2021
VersionString: 9.19.0
VersionCode: 901900000
NumDownloads: 17.429 M
Size: 35.625 MB
Offer: 0.00 GBP

> googleplay -a air.ITVMobilePlayer -v 901900000
GET https://play.googleapis.com/download/by-token/download?token=AOTCm0TiBZQdp...
~~~

1. https://support.google.com/googleplay/answer/7431675

## How to determine required features?

Use a command like this:

~~~
aapt dump badging file.apk
~~~

or check the `cmd/badging` folder.

## How to get Protocol Buffer fields?

Check Google Play Store (`com.android.vending`) with JADX, with the last
working version (2016):

~~~
versionCode='80441400' versionName='6.1.14'
~~~

- https://apkmirror.com/apk/google-inc/google-play-store
- https://github.com/whyorean/GPlayApi/blob/master/src/main/proto/GooglePlay.proto

## How to install Android App Bundle?

Bash:

~~~
adb install-multiple *.apk
~~~

PowerShell:

~~~
adb install-multiple (Get-ChildItem *.apk)
~~~

https://developer.android.com/guide/app-bundle/app-bundle-format

## How to install expansion file?

~~~
adb shell mkdir -p /sdcard/Android/obb/com.PirateBayGames.ZombieDefense2

adb push main.41.com.PirateBayGames.ZombieDefense2.obb `
/sdcard/Android/obb/com.PirateBayGames.ZombieDefense2/
~~~

https://developer.android.com/google/play/expansion-files

## INSTALL\_FAILED\_NO\_MATCHING\_ABIS

This can happen when trying to install ARM app on `x86`. If the APK is
`armeabi-v7a`, then Android 9 (API 28) will work. Also the emulator should be
`x86`. If the APK is `arm64-v8a`, then Android 11 (API 30) will work. Also the
emulator should be `x86_64`.

- https://android.stackexchange.com/questions/222094/install-failed
- https://stackoverflow.com/questions/36414219/install-failed-no-matching-abis

However note that this will still fail in some cases:

https://issuetracker.google.com/issues/207399356

## Version history

If you know the `versionCode`, you can get older APK [1]. Here is one from 2014:

~~~
googleplay -a com.google.android.youtube -v 5110
~~~

but I dont know how to get the old version codes, other than looking at
websites [2] that host the APKs.

1. https://android.stackexchange.com/questions/163181/how-to-download-old-app
2. https://apkmirror.com/uploads?appcategory=youtube

## Will you add features?

I created this tool for a very limited use (getting latest version string for
`com.google.android.youtube`). I am happy to fix bugs or add support for apps,
but I am not really interested in adding any features at this time. However if
you want to create an issue to make a suggestion, go ahead, but know that it
will probably not be implemented. However the module itself is open source [1],
so people can easily make their own tools using the module.

If you insist that some feature be implemented, I am willing to implement
features for people who are willing to make a donation to me. If that is your
situation, make sure to mention that in any communication. Minimum donation is
99 USD.

1. https://godocs.io/github.com/89z/googleplay
