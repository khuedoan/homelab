FROM nixos/nix

# TODO optimize Nix build speed (mount .nix at run time?)
COPY default.nix /default.nix

RUN nix-env --install --file /default.nix

CMD [ "nix-shell", "/default.nix" ]
