package googleplay

import (
   "fmt"
   "os"
   "testing"
)

func TestDelivery(t *testing.T) {
   home, err := os.UserHomeDir()
   if err != nil {
      t.Fatal(err)
   }
   token, err := OpenToken(home + "/googleplay/token.txt")
   if err != nil {
      t.Fatal(err)
   }
   device, err := OpenDevice(home + "/googleplay/x86.txt")
   if err != nil {
      t.Fatal(err)
   }
   id, err := device.AndroidID()
   if err != nil {
      t.Fatal(err)
   }
   head, err := token.Header(id, false)
   if err != nil {
      t.Fatal(err)
   }
   del, err := head.Delivery("com.google.android.youtube", 1524221376)
   if err != nil {
      t.Fatal(err)
   }
   fmt.Printf("%+v\n", del)
}
