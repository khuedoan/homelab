# DNS setup

Before you can access the home page at <https://home.example.com>, you'll need to update your DNS config.
Because everyone DNS setup are different, DNS automation is not in the scope of the project.

Some options for DNS config (choose one):

- Use [nip.io](https://nip.io) (suitable for a test environment)
- Change the DNS config in your router
- Change the DNS config at your domain registrar (doesn't work with the [`home.arpa`](https://datatracker.ietf.org/doc/html/rfc8375) domain)

Before continuing to the next section for some examples, run this command to get a list of subdomain and its address:

```sh
./scripts/get-dns-config
```

## Use nip.io

Preconfigured in the `dev` branch.

## In your router

You can add each subdomain one by one like the previous method, or use a wildcard `*.example.com` and point it to the IP address of the load blancer (from the output of the previous command)

## At your domain registrar

I'm using Cloudflare for DNS, continue to the next section for more information.
