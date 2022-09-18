# Alternate DNS setup

!!! info

        Skip this step if you already use the included Cloudflare setup

Before you can access the home page at <https://home.example.com>, you'll need to update your DNS config.

Some options for DNS config (choose one):

- Change the DNS config at your domain registrar (already included and automated)
- Change the DNS config in your router (also works if you don't own a domain)
- Use [nip.io](https://nip.io) (suitable for a test environment)

## At your domain registrar (recommended)

The default configuration is for Cloudflare DNS, but you can change the code to use other providers.

## In your router

!!! tip

    If you don't have a domain, you can use the `home.arpa` domain (according to [RFC-8375](https://datatracker.ietf.org/doc/html/rfc8375)).

You can add each subdomain one by one, or use a wildcard `*.example.com` and point it to the IP address of the load balancer.
To acquire a list of subdomains and their addresses, use this command:

```sh
./scripts/get-dns-config
```

## Use [nip.io](https://nip.io)

Preconfigured in the `dev` branch.
