package googleplay

import (
   "os"
   "testing"
   "time"
)

func TestToken(t *testing.T) {
   token, err := NewToken(email, password)
   if err != nil {
      t.Fatal(err)
   }
   home, err := os.UserHomeDir()
   if err != nil {
      t.Fatal(err)
   }
   if err := token.Create(home + "/googleplay/token.txt"); err != nil {
      t.Fatal(err)
   }
}

func TestHeader(t *testing.T) {
   home, err := os.UserHomeDir()
   if err != nil {
      t.Fatal(err)
   }
   token, err := OpenToken(home + "/googleplay/token.txt")
   if err != nil {
      t.Fatal(err)
   }
   for i := 0; i < 9; i++ {
      head, err := token.Header(0, false)
      if err != nil {
         t.Fatal(err)
      }
      if head.Auth == "" {
         t.Fatalf("%+v", head)
      }
      time.Sleep(time.Second)
   }
}
