
Self-signed cert with OpenSSL

```
/usr/local/opt/openssl/bin/openssl req -x509 -newkey rsa:2048 -out cert-self.pem -keyout key-self.pem -days 365 -nodes

/usr/local/opt/openssl/bin/openssl pkcs12 -in cert-self.pem -inkey key-self.pem -out cert-self.p12 -export
```
