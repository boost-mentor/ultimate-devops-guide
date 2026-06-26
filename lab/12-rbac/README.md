# RBAC practice

```bash
kubectl apply -f dev-reader-rbac.yaml
kubectl auth can-i get pods --as=system:serviceaccount:dev:dev-reader -n dev
kubectl auth can-i delete pods --as=system:serviceaccount:dev:dev-reader -n dev
kubectl auth can-i get pods --as=system:serviceaccount:dev:dev-reader -A
```

Вывод: Role работает внутри namespace. Для всего кластера нужен ClusterRole/ClusterRoleBinding.
