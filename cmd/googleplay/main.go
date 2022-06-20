package main

import (
   "flag"
   "github.com/89z/googleplay"
   "os"
   "path/filepath"
   "strings"
)

func main() {
   // a
   var app string
   flag.StringVar(&app, "a", "", "app")
   // d
   dir, err := os.UserHomeDir()
   if err != nil {
      panic(err)
   }
   dir = filepath.Join(dir, "googleplay")
   flag.StringVar(&dir, "d", dir, "user dir")
   // date
   var parse bool
   flag.BoolVar(&parse, "date", false, "parse date")
   // device
   var device bool
   flag.BoolVar(&device, "device", false, "create device")
   // email
   var email string
   flag.StringVar(&email, "email", "", "your email")
   // log
   flag.IntVar(&googleplay.Log.Level, "log", 0, "log level")
   // p
   var platformID int64
   flag.Int64Var(&platformID, "p", 0, googleplay.Platforms.String())
   // password
   var password string
   flag.StringVar(&password, "password", "", "your password")
   // purchase
   var (
      buf strings.Builder
      purchase bool
   )
   buf.WriteString("Purchase app. ")
   buf.WriteString("Only needs to be done once per Google account.")
   flag.BoolVar(&purchase, "purchase", false, buf.String())
   // s
   var single bool
   flag.BoolVar(&single, "s", false, "single APK")
   // v
   var version uint64
   flag.Uint64Var(&version, "v", 0, "app version")
   flag.Parse()
   if email != "" {
      err := do_token(dir, email, password)
      if err != nil {
         panic(err)
      }
   } else {
      platform := googleplay.Platforms[platformID]
      if device {
         err := do_device(dir, platform)
         if err != nil {
            panic(err)
         }
      } else if app != "" {
         head, err := do_header(dir, platform, single)
         if err != nil {
            panic(err)
         }
         if purchase {
            err := head.Purchase(app)
            if err != nil {
               panic(err)
            }
         } else if version >= 1 {
            err := do_delivery(head, app, version)
            if err != nil {
               panic(err)
            }
         } else {
            err := do_details(head, app, parse)
            if err != nil {
               panic(err)
            }
         }
      } else {
         flag.Usage()
      }
   }
}
