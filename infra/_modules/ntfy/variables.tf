variable "ntfy" {
  type = object({
    url   = string
    topic = string
  })
  sensitive = true
}
