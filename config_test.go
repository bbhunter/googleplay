package googleplay

import (
   "os"
   "testing"
   "time"
)

func checkin(id int64) error {
   platform := Platforms[id]
   home, err := os.UserHomeDir()
   if err != nil {
      return err
   }
   device, err := Phone.Checkin(platform)
   if err != nil {
      return err
   }
   platform += ".txt"
   if err := device.Create(home + "/googleplay/" + platform); err != nil {
      return err
   }
   time.Sleep(Sleep)
   return nil
}

func TestCheckinArmeabi(t *testing.T) {
   err := checkin(1)
   if err != nil {
      t.Fatal(err)
   }
}

func TestCheckinArm64(t *testing.T) {
   err := checkin(2)
   if err != nil {
      t.Fatal(err)
   }
}

func TestCheckinX86(t *testing.T) {
   err := checkin(0)
   if err != nil {
      t.Fatal(err)
   }
}
