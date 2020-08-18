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

Create secret that holds AWS credentials:

```
# cat secret.yaml

apiVersion: v1
kind: Secret
metadata:
  name: aws-credentials
  namespace: awspca-issuer-system
data:
  accesskey: <base64 encoding of AWS access key>
  secretkey: <base64 encoding of AWS secret key>
  region: <base64 encoding of AWS region key>
  arn: <base64 encoding of AWS Private CA ARN>
```

Create secret:

```  
# kubectl apply -f secret.yaml
```

Create resource AWSPCAIssuer for our controller:

```
# cat issuer.yaml

apiVersion: certmanager.awspca/v1alpha2
kind: AWSPCAIssuer
metadata:
  name: awspca-issuer
  namespace: awspca-issuer-system
spec:
  provisioner:
    name: aws-credentials
    accesskeyRef:
      key: accesskey
    secretkeyRef:
      key: secretkey
    regionRef:
      key: region
    arnRef:
      key: arn
```

Apply this configuration:

```
# kubectl apply -f issuer.yaml

# kubectl describe AWSPCAIssuer -n awspca-issuer-system

Name:         awspca-issuer
Namespace:    awspca-issuer-system
Labels:       <none>
Annotations:  API Version:  certmanager.awspca/v1alpha2
Kind:         AWSPCAIssuer
...
Spec:
  Provisioner:
    Accesskey Ref:
      Key:  accesskey
    Arn Ref:
      Key:  arn
    Name:   aws-credentials
    Region Ref:
      Key:  region
    Secretkey Ref:
      Key:  secretkey
Status:
  Conditions:
    Last Transition Time:  2020-08-18T04:34:33Z
    Message:               AWSPCAIssuer verified and ready to sign certificates
    Reason:                Verified
    Status:                True
    Type:                  Ready
Events:
  Type    Reason    Age                    From                     Message
  ----    ------    ----                   ----                     -------
  Normal  Verified  8m22s (x2 over 8m22s)  awspcaissuer-controller  AWSPCAIssuer verified and ready to sign certificates
```

Now create certificate:

```
# cat certificate.yaml

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
  # Renew 1 hour before the certificate expiration
  renewBefore: 1h
  isCA: false
  # The reference to the step issuer
  issuerRef:
    group: certmanager.awspca
    kind: AWSPCAIssuer
    name: awspca-issuer
```

```
# kubectl apply -f certificate.yaml
# kubectl describe Certificate backend-awspca -n awspca-issuer-system

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
    Last Transition Time:  2020-08-18T04:34:48Z
    Message:               Certificate is up to date and has not expired
    Reason:                Ready
    Status:                True
    Type:                  Ready
  Not After:               2020-08-19T04:34:45Z
  Not Before:              2020-08-18T03:34:45Z
  Renewal Time:            2020-08-19T03:34:45Z
  Revision:                1
Events:
  Type    Reason     Age    From          Message
  ----    ------     ----   ----          -------
  Normal  Issuing    6m1s   cert-manager  Issuing certificate as Secret does not exist
  Normal  Generated  6m     cert-manager  Stored new private key in temporary Secret resource "backend-awspca-7m9sx"
  Normal  Requested  6m     cert-manager  Created new CertificateRequest resource "backend-awspca-m2gz5"
  Normal  Issuing    5m51s  cert-manager  The certificate has been successfully issued
```

Check certificate and private key are present in secrets:                                             

```
# kubectl describe secrets backend-awspca-tls -n awspca-issuer-system   

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
tls.crt:  yyyy bytes
```
