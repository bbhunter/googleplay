package googleplay

import (
   "errors"
   "github.com/89z/format"
   "github.com/89z/format/protobuf"
   "net/http"
   "net/url"
   "strconv"
   "strings"
   "time"
)

func (h Header) Details(app string) (*Details, error) {
   req, err := http.NewRequest(
      "GET", "https://android.clients.google.com/fdfe/details", nil,
   )
   if err != nil {
      return nil, err
   }
   // half of the apps I test require User-Agent,
   // so just set it for all of them
   h.SetAgent(req.Header)
   h.SetAuth(req.Header)
   h.SetDevice(req.Header)
   req.URL.RawQuery = "doc=" + url.QueryEscape(app)
   LogLevel.Dump(req)
   res, err := new(http.Transport).RoundTrip(req)
   if err != nil {
      return nil, err
   }
   if res.StatusCode != http.StatusOK {
      return nil, errors.New(res.Status)
   }
   responseWrapper := make(protobuf.Message)
   responseWrapper.ReadFrom(res.Body)
   // .payload.detailsResponse.docV2
   docV2 := responseWrapper.Get(1).Get(2).Get(4)
   var det Details
   // The following fields will fail with wrong ABI, so try them first. If the
   // first one passes, then use native error for the rest.
   // .details.appDetails.versionCode
   det.VersionCode, err = docV2.Get(13).Get(1).GetVarint(3)
   if err != nil {
      return nil, errVersionCode{app}
   }
   // .details.appDetails.versionString
   det.VersionString, err = docV2.Get(13).Get(1).GetString(4)
   if err != nil {
      return nil, err
   }
   // .details.appDetails.installationSize
   det.Size, err = docV2.Get(13).Get(1).GetVarint(9)
   if err != nil {
      return nil, err
   }
   // .details.appDetails.uploadDate
   det.UploadDate, err = docV2.Get(13).Get(1).GetString(16)
   if err != nil {
      return nil, err
   }
   // .details.appDetails.file
   for _, file := range docV2.Get(13).Get(1).GetMessages(17) {
      // .fileType
      typ, err := file.GetVarint(1)
      if err != nil {
         return nil, err
      }
      det.File = append(det.File, typ)
   }
   // The following fields should work with any ABI.
   // .title
   det.Title, err = docV2.GetString(5)
   if err != nil {
      return nil, err
   }
   // .creator
   det.Creator, err = docV2.GetString(6)
   if err != nil {
      return nil, err
   }
   // .offer.micros
   det.Micros, err = docV2.Get(8).GetVarint(1)
   if err != nil {
      return nil, err
   }
   // .offer.currencyCode
   det.CurrencyCode, err = docV2.Get(8).GetString(2)
   if err != nil {
      return nil, err
   }
   // I dont know the name of field 70
   // .details.appDetails
   det.NumDownloads, err = docV2.Get(13).Get(1).GetVarint(70)
   if err != nil {
      return nil, err
   }
   return &det, nil
}

type errVersionCode struct {
   app string
}

func (e errVersionCode) Error() string {
   var buf strings.Builder
   buf.WriteString(e.app)
   buf.WriteString(" versionCode missing\n")
   buf.WriteString("Check nativePlatform")
   return buf.String()
}
type Details struct {
   Title string
   Creator string
   UploadDate string // Jun 1, 2021
   VersionString string
   VersionCode uint64
   NumDownloads uint64
   Size uint64
   File []uint64
   Micros uint64
   CurrencyCode string
}

func (d Details) String() string {
   var buf []byte
   buf = append(buf, "Title: "...)
   buf = append(buf, d.Title...)
   buf = append(buf, "\nCreator: "...)
   buf = append(buf, d.Creator...)
   buf = append(buf, "\nUploadDate: "...)
   buf = append(buf, d.UploadDate...)
   buf = append(buf, "\nVersionString: "...)
   buf = append(buf, d.VersionString...)
   buf = append(buf, "\nVersionCode: "...)
   buf = strconv.AppendUint(buf, d.VersionCode, 10)
   buf = append(buf, "\nNumDownloads: "...)
   buf = append(buf, format.LabelNumber(d.NumDownloads)...)
   buf = append(buf, "\nSize: "...)
   buf = append(buf, format.LabelSize(d.Size)...)
   buf = append(buf, "\nFile:"...)
   for _, file := range d.File {
      if file == 0 {
         buf = append(buf, " APK"...)
      } else {
         buf = append(buf, " OBB"...)
      }
   }
   buf = append(buf, "\nOffer: "...)
   buf = strconv.AppendUint(buf, d.Micros, 10)
   buf = append(buf, ' ')
   buf = append(buf, d.CurrencyCode...)
   return string(buf)
}

// This only works with English. You can force English with:
// Accept-Language: en
func (d Details) Time() (time.Time, error) {
   return time.Parse("Jan 2, 2006", d.UploadDate)
}
