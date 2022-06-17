package googleplay

import (
   "fmt"
   "os"
   "strconv"
   "testing"
   "time"
)

var apps = []appType{
   {"2021-12-08 00:00:00 +0000 UTC",0,"com.amctve.amcfullepisodes"},
   {"2022-02-02 00:00:00 +0000 UTC",2,"com.illumix.fnafar"},
   {"2022-02-14 00:00:00 +0000 UTC",0,"org.videolan.vlc"},
   {"2022-03-17 00:00:00 +0000 UTC",0,"com.google.android.apps.walletnfcrel"},
   {"2022-03-24 00:00:00 +0000 UTC",0,"app.source.getcontact"},
   {"2022-03-24 00:00:00 +0000 UTC",1,"com.miui.weather2"},
   {"2022-04-28 00:00:00 +0000 UTC",2,"com.miHoYo.GenshinImpact"},
   {"2022-05-11 00:00:00 +0000 UTC",1,"com.supercell.brawlstars"},
   {"2022-05-12 00:00:00 +0000 UTC",0,"com.clearchannel.iheartradio.controller"},
   {"2022-05-23 00:00:00 +0000 UTC",0,"kr.sira.metal"},
   {"2022-05-23 00:00:00 +0000 UTC",2,"com.kakaogames.twodin"},
   {"2022-05-30 00:00:00 +0000 UTC",1,"com.madhead.tos.zh"},
   {"2022-05-31 00:00:00 +0000 UTC",1,"com.xiaomi.smarthome"},
   {"2022-06-02 00:00:00 +0000 UTC",0,"org.thoughtcrime.securesms"},
   {"2022-06-02 00:00:00 +0000 UTC",1,"com.binance.dev"},
   {"2022-06-08 00:00:00 +0000 UTC",1,"com.sygic.aura"},
   {"2022-06-12 00:00:00 +0000 UTC",0,"br.com.rodrigokolb.realdrum"},
   {"2022-06-13 00:00:00 +0000 UTC",0,"com.app.xt"},
   {"2022-06-13 00:00:00 +0000 UTC",0,"com.google.android.youtube"},
   {"2022-06-13 00:00:00 +0000 UTC",0,"com.instagram.android"},
   {"2022-06-13 00:00:00 +0000 UTC",1,"com.axis.drawingdesk.v3"},
   {"2022-06-14 00:00:00 +0000 UTC",0,"com.pinterest"},
}

func TestDetails(t *testing.T) {
   home, err := os.UserHomeDir()
   if err != nil {
      t.Fatal(err)
   }
   token, err := OpenToken(home + "/googleplay/token.txt")
   if err != nil {
      t.Fatal(err)
   }
   head, err := token.Header(0, false)
   if err != nil {
      t.Fatal(err)
   }
   for _, app := range apps {
      platform := Platforms[app.platform]
      device, err := OpenDevice(home + "/googleplay/" + platform + ".txt")
      if err != nil {
         t.Fatal(err)
      }
      head.AndroidID, err = device.AndroidID()
      if err != nil {
         t.Fatal(err)
      }
      det, err := head.Details(app.id)
      if err != nil {
         t.Fatal(err)
      }
      if det.CurrencyCode == "" {
         t.Fatal(det)
      }
      if det.NumDownloads == 0 {
         t.Fatal(det)
      }
      if det.Size == 0 {
         t.Fatal(det)
      }
      if det.Title == "" {
         t.Fatal(det)
      }
      if det.VersionCode == 0 {
         t.Fatal(det)
      }
      if det.VersionString == "" {
         t.Fatal(det)
      }
      if det.UploadDate == "" {
         t.Fatal(det)
      }
      date, err := det.Time()
      if err != nil {
         t.Fatal(err)
      }
      app.date = date.String()
      fmt.Print(app, ",\n")
      time.Sleep(99 * time.Millisecond)
   }
}

func (a appType) String() string {
   var b []byte
   b = append(b, '{')
   b = strconv.AppendQuote(b, a.date)
   b = append(b, ',')
   b = strconv.AppendInt(b, a.platform, 10)
   b = append(b, ',')
   b = strconv.AppendQuote(b, a.id)
   b = append(b, '}')
   return string(b)
}

type appType struct {
   date string
   platform int64 // X-DFE-Device-ID
   id string
}
