# Safari Push Notifications

How to use this example in development.

1. Create a Website Push ID in [Apple's Member Center](https://developer.apple.com/membercenter/index.action). This requires a CSR from your keychain.

2. Download the certificate from Apple's website and add it to your Keychain.

3. Export a `.p12` with the certificate and the private key you generated during step 1.

4. Update main.go with the PushID you used in step 1.

5. Download and install [ngrok](https://ngrok.com/) and create an account.

6. Create a secure tunnel to port 5000 on your local machine by running `ngrok http 5000`.

7. You can visit [localhost:4040](http://localhost:4040) in your web browser to see requests as they happen.

8. In ngrok you will see a custom URL like `https://<something>.ngrok.io`. Update AllowedDomains, URLFormatString, and WebServiceURL in main.go based on your custom URL. Be sure to use HTTPS.

9. Build and run the example with `go run main.go -c /path/to/certificate.p12`

10. Visit your `https://<something>.ngrok.io` URL in Safari to request permission and then send a push notification, which will appear in your Notification Center.
