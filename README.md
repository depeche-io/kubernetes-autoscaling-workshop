

Intial setup
===

Use any standard K8s distribution, for example Kind:
- https://kind.sigs.k8s.io/docs/user/quick-start/

```
kind create cluster container-days
```


Bootstrap with ArgoCD Autopilot:
https://argocd-autopilot.readthedocs.io/en/stable/Installation-Guide/

argocd-autopilot repo bootstrap --recover --provider github --repo https://github.com/depeche-io/kubernetes-autoscaling-workshop.git/gitops/ --git-token XXX








kubectl port-forward -n argocd svc/argocd-server 8080:80

