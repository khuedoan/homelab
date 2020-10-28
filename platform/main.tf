resource "openstack_images_image_v2" "centos8" {
  name             = "CentOS_8"
  image_source_url = "https://cloud.centos.org/centos/8/x86_64/images/CentOS-8-GenericCloud-8.2.2004-20200611.2.x86_64.qcow2"
  container_format = "bare"
  disk_format      = "qcow2"
}
