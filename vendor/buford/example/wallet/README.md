# Wallet sign pass

How to sign a sample pass.

1. Download the PassKit support materials from the
 [Wallet Developer Guide](https://developer.apple.com/library/prerelease/ios/documentation/UserExperience/Conceptual/PassKit_PG/index.html) website.

2. Unzip WalletCompanionFiles.zip and copy `SamplePasses/Event.pass` into this folder.

3. Register a Pass Type ID in [Apple's Member Center](https://developer.apple.com/membercenter/index.action). This requires a CSR from your keychain.

4. Download the `.cer` and add it to Keychain. Then export a `.p12` with the certificate and the private key you generated during step 3.

5. Modify the passTypeIdentifier and teamIdentifier in pass.json. The teamIdentifier is the Organizational Unit in the .cer file.

6. Download [Apple Intermediate Certificate](http://www.apple.com/certificateauthority/) WWDR Certificate (Expiring 02/07/23)

7. Build and run the example with `go run main.go -c /path/to/certificate.p12 -i /path/to/AppleWWDRCA.cer`
