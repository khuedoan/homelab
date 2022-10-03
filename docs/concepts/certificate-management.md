# Certificate management

Certificates are generated and managed by [cert-manager](https://cert-manager.io) with [Let's Encrypt](https://letsencrypt.org).
By default certificates are valid for 90 days and will be renewed after 60 days.

cert-manager watches `Ingress` resources across the cluster. When you create an `Ingress` with a [supported annotation](https://cert-manager.io/docs/usage/ingress/#supported-annotations):

```yaml hl_lines="5 13 14"
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
  name: foo
spec:
  rules:
    - host: foo.example.com
      # ...
  tls:
    - hosts:
        - foo.example.com
      secretName: foo-tls-certificate
```

```mermaid
flowchart LR
  User -- 6 --> Ingress

  subgraph cluster[Homelab cluster]
    Ingress --- Secret
    Ingress -. 1 .-> Certificate
    Certificate -. 5 .-> Secret
    Certificate -- 2 --> CertificateRequest -- 3 --> Order -- 4 --> Challenge
  end

  Order -.- ACMEServer[ACME server]

  subgraph dnsprovider[DNS provider]
    TXT
  end

  Challenge -- 4.a --> TXT
  ACMEServer -.- Challenge
  ACMEServer -. 4.b .-> TXT
```

1. cert-manager creates a corresponding `Certificate` resources
2. Based on the `Certificate` resource, cert-manager creates a `CertificateRequest` resource to request a signed certificate from the configured `ClusterIssuer`
3. The `CertificateRequest` will create an order with an ACME server (we use Let's Encrypt), which is represented by the `Order` resource
4. Then cert-manager will perform a [DNS-01](https://cert-manager.io/docs/configuration/acme/dns01) `Challenge`:
    1. Create a DNS TXT record (contains a computed key)
    2. The ACME server retrieve this key via a DNS lookup and validate that we own the domain for the requested certificate
7. cert-manager stores the certificate (typically `tls.crt` and `tls.key`) in the `Secret` specified in the `Ingress` configuration
8. Now you can access the HTTPS website with a valid certificate

A much more detailed diagram can be found in the official documentation under [certificate lifecycle](https://cert-manager.io/docs/concepts/certificate/#certificate-lifecycle).
