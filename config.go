package googleplay

import (
   "bytes"
   "github.com/89z/format"
   "github.com/89z/format/protobuf"
   "net/http"
   "os"
   "strconv"
)

func (d Device) Create(name string) error {
   file, err := format.Create(name)
   if err != nil {
      return err
   }
   defer file.Close()
   if _, err := d.WriteTo(file); err != nil {
      return err
   }
   return nil
}

func OpenDevice(name string) (*Device, error) {
   file, err := os.Open(name)
   if err != nil {
      return nil, err
   }
   defer file.Close()
   var dev Device
   dev.Message = make(protobuf.Message)
   if _, err := dev.ReadFrom(file); err != nil {
      return nil, err
   }
   return &dev, nil
}

type NativePlatform map[int64]string

var Platforms = NativePlatform{
   // com.google.android.youtube
   0: "x86",
   // com.miui.weather2
   1: "armeabi-v7a",
   // com.kakaogames.twodin
   2: "arm64-v8a",
}

func (n NativePlatform) String() string {
   var buf []byte
   buf = append(buf, "nativePlatform"...)
   for key, val := range n {
      buf = append(buf, '\n')
      buf = strconv.AppendInt(buf, key, 10)
      buf = append(buf, ": "...)
      buf = append(buf, val...)
   }
   return string(buf)
}

// These can use default values, but they must all be included
type Config struct {
   DeviceFeature []string
   GlEsVersion uint64
   GlExtension []string
   HasFiveWayNavigation uint64
   HasHardKeyboard uint64
   Keyboard uint64
   Navigation uint64
   ScreenDensity uint64
   ScreenLayout uint64
   SystemSharedLibrary []string
   TouchScreen uint64
}

var Phone = Config{
   DeviceFeature: []string{
      // app.source.getcontact
      "android.hardware.location.gps",
      // br.com.rodrigokolb.realdrum
      "android.software.midi",
      // com.app.xt
      "android.hardware.camera.front",
      // com.clearchannel.iheartradio.controller
      "android.hardware.microphone",
      // com.google.android.apps.walletnfcrel
      "android.software.device_admin",
      // com.google.android.youtube
      "android.hardware.touchscreen",
      "android.hardware.wifi",
      // com.illumix.fnafar
      "android.hardware.sensor.gyroscope",
      // com.madhead.tos.zh
      "android.hardware.sensor.accelerometer",
      // com.miHoYo.GenshinImpact
      "android.hardware.opengles.aep",
      // com.pinterest
      "android.hardware.camera",
      "android.hardware.location",
      "android.hardware.screen.portrait",
      // com.supercell.brawlstars
      "android.hardware.touchscreen.multitouch",
      // com.sygic.aura
      "android.hardware.location.network",
      // com.xiaomi.smarthome
      "android.hardware.bluetooth",
      "android.hardware.bluetooth_le",
      "android.hardware.camera.autofocus",
      "android.hardware.usb.host",
      // kr.sira.metal
      "android.hardware.sensor.compass",
      // org.thoughtcrime.securesms
      "android.hardware.telephony",
      // org.videolan.vlc
      "android.hardware.screen.landscape",
   },
   SystemSharedLibrary: []string{
      // com.amctve.amcfullepisodes
      "org.apache.http.legacy",
      // com.binance.dev
      "android.test.runner",
      // com.miui.weather2
      "global-miui11-empty.jar",
   },
   // com.axis.drawingdesk.v3
   GlEsVersion: 0x9_9999,
   GlExtension: []string{
      // com.instagram.android
      "GL_OES_compressed_ETC1_RGB8_texture",
      // com.kakaogames.twodin
      "GL_KHR_texture_compression_astc_ldr",
   },
}

func (d Device) AndroidID() (uint64, error) {
   return d.GetFixed64(7)
}

// A Sleep is needed after this.
func (c Config) Checkin(platform string) (*Device, error) {
   checkin := protobuf.Message{
      4: protobuf.Message{ // checkin
         1: protobuf.Message{ // build
            10: protobuf.Varint(29), // sdkVersion
         },
      },
      14: protobuf.Varint(3), // version
      18: protobuf.Message{ // deviceConfiguration
         1: protobuf.Varint(c.TouchScreen),
         2: protobuf.Varint(c.Keyboard),
         3: protobuf.Varint(c.Navigation),
         4: protobuf.Varint(c.ScreenLayout),
         5: protobuf.Varint(c.HasHardKeyboard),
         6: protobuf.Varint(c.HasFiveWayNavigation),
         7: protobuf.Varint(c.ScreenDensity),
         8: protobuf.Varint(c.GlEsVersion),
         11: protobuf.String(platform), // nativePlatform
      },
   }
   for _, library := range c.SystemSharedLibrary {
      // .deviceConfiguration.systemSharedLibrary
      checkin.Get(18).AddString(9, library)
   }
   for _, extension := range c.GlExtension {
      // .deviceConfiguration.glExtension
      checkin.Get(18).AddString(15, extension)
   }
   for _, name := range c.DeviceFeature {
      // .deviceConfiguration.deviceFeature
      checkin.Get(18).Add(26, protobuf.Message{
         1: protobuf.String(name),
      })
   }
   buf := new(bytes.Buffer)
   checkin.WriteTo(buf)
   req, err := http.NewRequest(
      "POST", "https://android.googleapis.com/checkin", buf,
   )
   if err != nil {
      return nil, err
   }
   req.Header.Set("Content-Type", "application/x-protobuffer")
   LogLevel.Dump(req)
   res, err := new(http.Transport).RoundTrip(req)
   if err != nil {
      return nil, err
   }
   defer res.Body.Close()
   var dev Device
   dev.Message = make(protobuf.Message)
   dev.ReadFrom(res.Body)
   return &dev, nil
}

type Device struct {
   protobuf.Message
}
