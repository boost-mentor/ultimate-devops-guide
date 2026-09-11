# Kubespray v2.31.0: поддерживаемые версии для этой лабораторной

Источник: официальный репозиторий `kubernetes-sigs/kubespray`, tag `v2.31.0`,
commit `1c9add48975060f45396b34d8e022c30d7f80dab`.

- В README этого release Kubernetes `1.35.4` указан среди поддерживаемых
  компонентов.
- В checksum-таблице release присутствуют и исходная `1.34.7`, и целевая
  `1.35.4`.
- Поэтому переход `1.34.7 → 1.35.4` остаётся внутри одной соседней minor-линии
  и внутри матрицы одного закреплённого Kubespray release.

Проверять на экране в официальном clone:

- `README.md`, раздел `Supported Components`;
- `roles/kubespray_defaults/vars/main/checksums.yml`;
- `roles/kubespray_defaults/defaults/main/main.yml`.

Ссылка: <https://github.com/kubernetes-sigs/kubespray/tree/v2.31.0>
