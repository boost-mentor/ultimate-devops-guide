# Лабораторная 1.1 — Terraform без облака

Цель лабораторной — увидеть полный цикл Terraform на безопасных локальных
файлах, прежде чем создавать сеть и виртуальные машины в облаке.

Открывать файлы в таком порядке:

1. `versions.tf` — версия Terraform и источник provider `local`;
2. `providers.tf` — подключение provider;
3. `variables.tf` — входные значения и их проверки;
4. `env/demo.tfvars` — конкретные значения этого запуска;
5. `locals.tf` — внутренние вычисления проекта;
6. `templates/` — шаблоны содержимого файлов;
7. `resources.tf` — три управляемых файла;
8. `outputs.tf` — пути, которые Terraform отдаёт наружу.

`terraform init` читает блок `required_providers`, скачивает provider `local` и
записывает выбранную версию в lock-файл. Он ещё не создаёт ресурсы.
`terraform validate` проверяет синтаксис и связи внутри конфигурации; например,
ссылка на несуществующую переменную остановит проверку до любого `apply`.
`terraform plan -var-file=env/demo.tfvars` подставляет значения именно из этого
файла и сравнивает код, state и текущие файлы. Plan только показывает действия.

Полный наблюдаемый цикл:

```bash
terraform init
terraform validate
terraform plan -var-file=env/demo.tfvars
terraform apply -var-file=env/demo.tfvars
terraform state list
terraform plan -var-file=env/demo.tfvars
```

После `apply` физический результат лежит в `build/`, `terraform state list`
показывает три управляемых адреса, а повторный `plan` должен ответить
`No changes`.

Опыт с drift — изменить управляемый файл вне Terraform:

```bash
printf '\nmanual_change=true\n' >> build/journal.txt
terraform plan -var-file=env/demo.tfvars
```

Provider `local` опознаёт файл по контрольной сумме содержимого, поэтому
изменённый файл перестаёт совпадать с управляемым экземпляром из state. На
Terraform 1.13.4 и provider `local` 2.5.2 план сразу показывает
`# local_file.journal will be created` и `1 to add`, без отдельной плашки
`Objects have changed outside of Terraform`. Файл физически не удалён — это
модель восстановления ресурса provider. Дальше два корректных пути: если
ручная правка ошибочна — `terraform apply -var-file=env/demo.tfvars` возвращает
описанное состояние, и контрольный `plan` снова отвечает `No changes`; если
правка нужна — сначала перенесите её в шаблон или переменные и добейтесь
пустого плана. Не удаляйте state и не запускайте `apply` вслепую, пока причина
расхождения не понята.

State не удаляется вручную. Учебные ресурсы удаляются только через
`terraform destroy -var-file=env/demo.tfvars`.
