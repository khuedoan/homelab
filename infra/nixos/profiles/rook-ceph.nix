{
  # Rook Ceph runs OSDs and the RBD CSI driver on the host kernel.
  boot.kernelModules = [
    "ceph"
    "rbd"
  ];
}
