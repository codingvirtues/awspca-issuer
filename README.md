# AWS Certificate Manager Private Certificate Authority

AWS Certificate Manager Private CA is a Certificate Authority managed by AWS (https://aws.amazon.com/certificate-manager/private-certificate-authority/). It allows creation of root and intermediate CA's that can issue certificates for entities blessed by the CA.

# cert-manager

cert-manager manages certificates in Kubernetes environment (among others) and keeps track of renewal requirements (https://cert-manager.io/). It supports various in-built issuers that issue the certificates to be managed by cert-manager.

# AWS Private CA Issuer

This project plugs into cert-manager as an external issuer that talks to AWS Certificate Manager Private CA to get certificates issued for your Kubernetes environment.

# Setup

Install cert-manager first (https://cert-manager.io/docs/installation/kubernetes/), version 0.16.1 or later.

Clone this repo and perform following steps to install controller:

```
# make build
# make docker
# make deploy
```

Create resource AWSPCAIssuer for the above controller:

# cat issuer.yaml
```
apiVersion: certmanager.awspca/v1alpha2
kind: AWSPCAIssuer
metadata:
  name: awspca-issuer
  namespace: awspca-issuer-system
spec:
  provisioner:
    accesskey: <aws access key>
    secretkey: <aws seecret key>
    region: <region>
    arn: <ARN of AWS Private CA>
```

# kubectl apply -f issuer.yaml

# kubectl describe AWSPCAIssuer -n awspca-issuer-system

```
Name:         awspca-issuer
Namespace:    awspca-issuer-system
Labels:       <none>
Annotations:  API Version:  certmanager.awspca/v1alpha2
Kind:         AWSPCAIssuer
...
Status:
  Conditions:
    Last Transition Time:  2020-08-08T19:40:26Z
    Message:               AWSPCAIssuer verified and ready to sign certificates
    Reason:                Verified
    Status:                True
    Type:                  Ready
Events:
  Type    Reason    Age                From                     Message
  ----    ------    ----               ----                     -------
  Normal  Verified  18m (x2 over 18m)  awspcaissuer-controller  AWSPCAIssuer verified and ready to sign certificates
```

Now create certificate:

# cat certificate.yaml
```
apiVersion: cert-manager.io/v1alpha2
kind: Certificate
metadata:
  name: backend-awspca
  namespace: awspca-issuer-system
spec:
  # The secret name to store the signed certificate
  secretName: backend-awspca-tls
  # Common Name
  commonName: foo.com
  # DNS SAN
  dnsNames:
    - localhost
    - foo.com
  # IP Address SAN
  ipAddresses:
    - "127.0.0.1"
  # Duration of the certificate
  duration: 24h
  # Renew 8 hours before the certificate expiration
  renewBefore: 1h
  isCA: false
  # The reference to the step issuer
  issuerRef:
    group: certmanager.awspca
    kind: AWSPCAIssuer
    name: awspca-issuer
```

# kubectl apply -f certificate.yaml
# kubectl describe Certificate backend-awspca -n awspca-issuer-system

```
Name:         backend-awspca
Namespace:    awspca-issuer-system
Labels:       <none>
Annotations:  API Version:  cert-manager.io/v1alpha3
Kind:         Certificate
...
Spec:
  Common Name:  foo.com
  Dns Names:
    localhost
    foo.com
  Duration:  24h0m0s
  Ip Addresses:
    127.0.0.1
  Issuer Ref:
    Group:       certmanager.awspca
    Kind:        AWSPCAIssuer
    Name:        awspca-issuer
  Renew Before:  1h0m0s
  Secret Name:   backend-awspca-tls
Status:
  Conditions:
    Last Transition Time:  2020-08-08T19:42:49Z
    Message:               Certificate is up to date and has not expired
    Reason:                Ready
    Status:                True
    Type:                  Ready
  Not After:               2020-08-09T19:42:46Z
Events:
  Type    Reason        Age   From          Message
  ----    ------        ----  ----          -------
  Normal  GeneratedKey  18m   cert-manager  Generated a new private key
  Normal  Requested     18m   cert-manager  Created new CertificateRequest resource "backend-awspca-3903941586"
  Normal  Issued        18m   cert-manager  Certificate issued successfully
```

Check certificate and private key are present in secrets:

# kubectl describe secrets backend-awspca-tls -n awspca-issuer-system                                                

```
Name:         backend-awspca-tls
Namespace:    awspca-issuer-system
Labels:       <none>
Annotations:  cert-manager.io/alt-names: localhost,foo.com
              cert-manager.io/certificate-name: backend-awspca
              cert-manager.io/common-name: foo.com
              cert-manager.io/ip-sans: 127.0.0.1
              cert-manager.io/issuer-kind: AWSPCAIssuer
              cert-manager.io/issuer-name: awspca-issuer
              cert-manager.io/uri-sans:

Type:  kubernetes.io/tls

Data
====
tls.key:  xxxx bytes
ca.crt:   yyyy bytes
tls.crt:  zzzz bytes
```
