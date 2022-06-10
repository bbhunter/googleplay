# GooglePlay

> Fear plays an interesting role in our lives. How dare we let it motivate us?
> How dare we let it into our decision-making, into our livelihoods, into our
> relationships?
>
> It's funny, isn't it, we take a day a year to dress up in costume and
> celebrate fear?
>
> [Spooked (The Office) (2011)](//youtube.com/watch?v=9Ex4UcLaYNc)

Download APK from Google Play or send API requests

## How to install?

This module works with Windows, macOS or Linux. First, [download Go][2] and
extract archive. Then [download&nbsp;GooglePlay][3] and extract archive. Then
navigate to `googleplay-master/cmd/googleplay`, and enter:

~~~
go build
~~~

[2]://go.dev/dl
[3]://github.com/89z/googleplay/archive/refs/heads/master.zip

## Tool examples

Before trying these examples, make sure the Google account you are using has
logged into the Play&nbsp;Store at least once before. Also you need to have
accepted the Google Play terms and conditions. Create a file containing token
(`aas_et`) for future requests:

~~~
googleplay -email EMAIL -password PASSWORD
~~~

Create a file containing `Android_ID` (GSF ID) for future requests:

~~~
googleplay -device
~~~

Get app details:

~~~
> googleplay -a com.google.android.youtube
Title: YouTube
Creator: Google LLC
UploadDate: 2022-05-12
VersionString: 17.19.34
VersionCode: 1529337280
NumDownloads: 11.822 B
Size: 46.727 MB
File: APK APK APK APK
Offer: 0 USD
~~~

Purchase app. Only needs to be done once per Google account:

~~~
googleplay -a com.google.android.youtube -purchase
~~~

Download APK. You need to specify any valid version code. The latest code is
provided by the previous details command. If APK is split, all pieces will be
downloaded:

~~~
googleplay -a com.google.android.youtube -v 1529337280
~~~

## Sponsor

I really need help financially, so if you are able, please donate using the
sponsor link. If you cannot use PayPal, let me know, and I can see about adding
other methods. Contact me with any business opportunities:

- Email srpen6@gmail.com
- Discord 89z#4149

## API

https://github.com/89z/googleplay
