# Media management

!!! warning

    This is for educational purposes only :wink: Use it at your own risk!

## Initial setup

- Jellyfin `https://jellyfin.example.com`:
    - Create an `admin` user and save the credentials to your password manager
    - Add media libraries:
        - Movies at `/media/movies`
        - Shows at `/media/shows`
- Radarr `https://radarr.example.com`:
    - Authentication method: Forms
    - Create an `admin` user and save the credentials to your password manager
    - Navigate to Settings -> Download Clients -> Add -> Transmission (you can keep the default address and port)
    - Navigate to Settings -> Media Management -> Add Root Folder `/movies`
    - Navigate to Settings -> General -> API Key -> copy it for the next steps (or save it to your password manager)
- Sonarr `https://sonarr.example.com`: same as Radarr but use `/shows` for the root folder
- Prowlarr `https://prowlarr.example.com`:
    - Authentication method: Forms
    - Create an `admin` user and save the credentials to your password manager
    - Navigate to Settings -> Apps -> Add:
        - Radarr: paste the API key (you can keep the default address and port)
        - Sonarr: same as Radarr
    - Go back to Indexers -> Add New Indexers
- Jellyseerr `https://jellyseerr.example.com`:
    - Sign In
        - Use your Jellyfin account
        - URL: `https://jellyfin.example.com`
        - Email: you can enter anything
        - Username: `admin`
        - Password: same as Jellyfin
    - Configure Media Server
        - Enable Movies and Shows
    - Configure Services:
        - Add Radarr Server:
            - Default Server: true
            - Name: Radarr
            - Hostname: localhost
            - Port: use default
            - API Key: from previous step
            - Click Test
            - Quality Profile: choose whatever suits you
            - Root folder: `/movies`
            - External URL: `https://radarr.example.com`
            - Enable Scan: true
        - Add Sonarr Server: similar to Radarr

Optionally, for convenience, you can add a `guest` account without a password in Jellyfin,
allow access to all libraries, and allow that account to manage requests in Jellyseerr.

## Usage

Here's a suggested flow:

- Users using the `guest` account can request media in Jellyseerr
- Admins approve the request (or you can enable auto-approve)
- Wait for the media to be downloaded
- Watch on Jellyfin

You may need to increase the volume size depending on your usage.
