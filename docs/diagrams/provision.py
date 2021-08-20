from diagrams import Cluster, Diagram
from diagrams.generic.os import LinuxGeneral
from diagrams.k8s.ecosystem import Helm, Kustomize
from diagrams.k8s.infra import Node
from diagrams.onprem.client import Client
from diagrams.onprem.compute import Server
from diagrams.onprem.container import Docker
from diagrams.onprem.gitops import ArgoCD
from diagrams.onprem.iac import Ansible, Terraform

graph_attr = {
    "pad": "0"
}

with Diagram("Provision (separated by logical layers)", graph_attr=graph_attr, outformat="jpg", show=False):
    controller = Client("Initial controller")
    bare_metal_machines = Server(f"Server(s)")

    with Cluster("./metal"):
        ansible = Ansible("Ansible")
        pxe = Docker("Ephemeral PXE server")
        os = LinuxGeneral(f"CoreOS")

        controller >> ansible
        ansible >> pxe
        pxe >> os
        os >> bare_metal_machines

        with Cluster("./cluster"):
            terraform = Terraform("RKE")
            kubernetes_nodes = Node("Kuberentes node(s)")
            argocd = ArgoCD("ArgoCD")

            controller >> terraform
            terraform >> kubernetes_nodes
            terraform >> argocd

            with Cluster("./apps"):
                kustomize = Kustomize("Kustomize")
                charts = Helm("Helm charts")

                controller >> kustomize
                kustomize >> argocd
                argocd >> charts
