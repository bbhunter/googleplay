package googleplay

import (
   "errors"
   "github.com/89z/format"
   "github.com/89z/format/crypto"
   "github.com/89z/format/net"
   "net/http"
   "net/url"
   "os"
   "strconv"
   "strings"
   "time"
)

func (t Token) Create(name string) error {
   file, err := format.Create(name)
   if err != nil {
      return err
   }
   defer file.Close()
   if _, err := t.WriteTo(file); err != nil {
      return err
   }
   return nil
}

func OpenToken(name string) (*Token, error) {
   file, err := os.Open(name)
   if err != nil {
      return nil, err
   }
   defer file.Close()
   var tok Token
   tok.Values = net.NewValues()
   if _, err := tok.ReadFrom(file); err != nil {
      return nil, err
   }
   return &tok, nil
}

func (t Token) Header(androidID uint64, single bool) (*Header, error) {
   // these values take from Android API 28
   body := url.Values{
      "Token": {t.Token()},
      "app": {"com.android.vending"},
      "client_sig": {"38918a453d07199354f8b19af05ec6562ced5788"},
      "service": {"oauth2:https://www.googleapis.com/auth/googleplay"},
   }.Encode()
   req, err := http.NewRequest(
      "POST", "https://android.googleapis.com/auth", strings.NewReader(body),
   )
   if err != nil {
      return nil, err
   }
   req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
   LogLevel.Dump(req)
   res, err := new(http.Transport).RoundTrip(req)
   if err != nil {
      return nil, err
   }
   defer res.Body.Close()
   if res.StatusCode != http.StatusOK {
      return nil, errors.New(res.Status)
   }
   var head Header
   head.SDK = 9
   head.AndroidID = androidID
   if single {
      head.VersionCode = 8091_9999 // single APK
   } else {
      head.VersionCode = 9999_9999
   }
   val := net.NewValues()
   val.ReadFrom(res.Body)
   head.Auth = val.Get("Auth")
   return &head, nil
}

const Sleep = 4 * time.Second

var LogLevel format.LogLevel

type Header struct {
   AndroidID uint64 // X-DFE-Device-ID
   SDK int64 // User-Agent
   VersionCode int64 // User-Agent
   Auth string // Authorization
}

func (h Header) SetAgent(head http.Header) {
   var buf []byte
   buf = append(buf, "Android-Finsky (sdk="...)
   buf = strconv.AppendInt(buf, h.SDK, 10)
   buf = append(buf, ",versionCode="...)
   buf = strconv.AppendInt(buf, h.VersionCode, 10)
   head.Set("User-Agent", string(buf))
}

func (h Header) SetAuth(head http.Header) {
   head.Set("Authorization", "Bearer " + h.Auth)
}

func (h Header) SetDevice(head http.Header) {
   device := strconv.FormatUint(h.AndroidID, 16)
   head.Set("X-DFE-Device-ID", device)
}

// Purchase app. Only needs to be done once per Google account.
func (h Header) Purchase(app string) error {
   query := "doc=" + url.QueryEscape(app)
   req, err := http.NewRequest(
      "POST", "https://android.clients.google.com/fdfe/purchase",
      strings.NewReader(query),
   )
   if err != nil {
      return err
   }
   h.SetAuth(req.Header)
   h.SetDevice(req.Header)
   req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
   LogLevel.Dump(req)
   res, err := new(http.Transport).RoundTrip(req)
   if err != nil {
      return err
   }
   defer res.Body.Close()
   if res.StatusCode != http.StatusOK {
      return errors.New(res.Status)
   }
   return nil
}

func (t Token) Token() string {
   return t.Get("Token")
}

// You can also use host "android.clients.google.com", but it also uses
// TLS fingerprinting.
func NewToken(email, password string) (*Token, error) {
   body := url.Values{
      "Email": {email},
      "Passwd": {password},
      "client_sig": {""},
      "droidguard_results": {""},
   }.Encode()
   req, err := http.NewRequest(
      "POST", "https://android.googleapis.com/auth", strings.NewReader(body),
   )
   if err != nil {
      return nil, err
   }
   req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
   hello, err := crypto.ParseJA3(crypto.AndroidAPI26)
   if err != nil {
      return nil, err
   }
   LogLevel.Dump(req)
   res, err := crypto.Transport(hello).RoundTrip(req)
   if err != nil {
      return nil, err
   }
   defer res.Body.Close()
   if res.StatusCode != http.StatusOK {
      return nil, errors.New(res.Status)
   }
   var tok Token
   tok.Values = net.NewValues()
   if _, err := tok.ReadFrom(res.Body); err != nil {
      return nil, err
   }
   return &tok, nil
}

type Token struct {
   net.Values
}
