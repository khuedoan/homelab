# DNS setup

Before you can access the home page at <https://home.example.com>, you'll need to update your DNS config.

Some options for DNS config (choose one):

- Change the DNS config at your domain registrar (easy to automate)
- Change the DNS config in your router (also works with the [`home.arpa`](https://datatracker.ietf.org/doc/html/rfc8375) domain)
- Use [nip.io](https://nip.io) (suitable for a test environment)

## At your domain registrar (recommended)

I'm using Cloudflare for DNS, continue to the next section for more information.

## In your router

You can add each subdomain one by one, or use a wildcard `*.example.com` and point it to the IP address of the load balancer.
To acquire a list of subdomains and their addresses, use this command:

```sh
./scripts/get-dns-config
```

## Use nip.io

Preconfigured in the `dev` branch.
