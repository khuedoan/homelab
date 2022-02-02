from diagrams import Cluster, Diagram
from diagrams.generic.os import LinuxGeneral
from diagrams.k8s.ecosystem import Helm
from diagrams.k8s.infra import Node
from diagrams.onprem.container import Docker
from diagrams.onprem.gitops import ArgoCD

graph_attr = {
    "pad": "0"
}

with Diagram("Provisioning flow", graph_attr=graph_attr, outformat="png", show=False):
    with Cluster("./metal"):
        pxe = Docker("PXE server")
        os = LinuxGeneral(f"Rocky Linux")
        cluster = Node("K3s")

        pxe >> os >> cluster

    with Cluster("./bootstrap"):
        argocd = ArgoCD("ArgoCD")

        cluster >> argocd

    with Cluster("./system"):
        system_charts = Helm("System charts")

        argocd >> system_charts

    with Cluster("./platform"):
        platform_charts = Helm("Platform charts")

        argocd >> platform_charts

    with Cluster("./apps"):
        app_charts = Helm("Application charts")

        argocd >> app_charts
