# Лабораторная 1.2 — небольшой Ansible-проект

Проект настраивает один учебный host — текущий ноутбук — и создаёт файл
`/tmp/devops-may-cry/dispatcher.conf`. Та же структура затем встречается в
Kubespray: inventory, group variables, host variables, playbook и roles.

Открывать файлы в таком порядке:

1. `ansible.cfg` — где Ansible ищет inventory и roles;
2. `inventory/hosts.ini` — группа `night_office` и host `localhost`;
3. `inventory/group_vars/all.yml` — общие значения;
4. `inventory/group_vars/night_office.yml` — значения группы;
5. `inventory/host_vars/localhost.yml` — значение конкретного host;
6. `roles/dispatcher/defaults/main.yml` — самые слабые defaults роли;
7. `roles/dispatcher/tasks/main.yml` — действия по порядку;
8. `roles/dispatcher/templates/dispatcher.conf.j2` — шаблон результата;
9. `roles/dispatcher/handlers/main.yml` — реакция только на изменение;
10. `playbooks/site.yml` — play, который связывает группу и role.

Первый запуск создаёт/изменяет файл. После чтения результата второй запуск
должен завершиться с `changed=0`. Это доказывает идемпотентность данного
сценария, но корректность содержимого доказывается отдельным чтением файла.

Проверки выполняются в том же порядке, в котором проект был разобран:

```bash
ansible-inventory --graph
ansible-playbook playbooks/site.yml --syntax-check
ansible-playbook playbooks/site.yml --list-tags
ansible-playbook playbooks/site.yml
ansible-playbook playbooks/site.yml
cat /tmp/devops-may-cry/dispatcher.conf
```

`ansible-inventory --graph` не рисует сетевой граф: он показывает, какие hosts
входят в какие inventory-группы. `--syntax-check` останавливает запуск при
ошибке YAML или структуре playbook. `--list-tags` показывает доступные
`config`/`verify`; в production tags часто позволяют выполнить только выбранную
часть роли, но они не заменяют корректный полный прогон.

Файл `playbooks/precedence.yml` — отдельный опыт. Он задаёт `gear` на уровне
play и позволяет увидеть, что более конкретное значение перекрывает defaults,
group_vars и host_vars. Секреты так не демонстрируются: их передают через Vault
или secret manager и не печатают через `debug`.

Если `changed=0`, это доказывает только отсутствие новой правки. Итоговая
приёмка читает созданный файл и проверяет ожидаемые значения. В удалённом
проекте к этому добавятся доступ по SSH, права, status сервиса и его реальный
ответ, а не только зелёный `PLAY RECAP`.

## Почему playbook идемпотентен не автоматически

Многие стандартные модули (`file`, `template`, `apt`, `service` со
`state: started`) сначала сравнивают текущее состояние с желаемым и ничего
не меняют, если цель уже достигнута. Но Ansible не превращает произвольный
playbook в идемпотентный сам по себе: результат зависит от выбранного модуля,
его параметров и дополнительной логики автора. `command`/`shell`/`raw`
желаемого состояния не знают и по умолчанию всегда отчитываются `changed`,
а `state: restarted` — действие при каждом прогоне, а не требование.

Файл `non_idempotent_examples.yml` — справочник таких примеров: плохой
`shell` с `>>` и три исправления (`lineinfile`, `creates`/`removes`,
честный `changed_when`). Этот файл НЕ запускается в основном flow
лабораторной и не подключён к `site.yml` — он читается как конспект.

## Drift-lab: ручное расхождение и восстановление

```bash
printf '\nmanual_override=true\n' \
  >> /tmp/devops-may-cry/dispatcher.conf
ansible-playbook playbooks/site.yml --check --diff
ansible-playbook playbooks/site.yml --diff
cat /tmp/devops-may-cry/dispatcher.conf
ansible-playbook playbooks/site.yml
```

`--check` не меняет систему: модули сообщают, было бы изменение. `--diff`
показывает before/after для файлов. Поддержка обоих режимов зависит от
модуля: например, `command` в check-режиме по умолчанию пропускается.
Restore-прогон переписывает файл из шаблона — на нём срабатывает handler,
потому что template реально изменил файл. Контрольный прогон снова даёт
`changed=0`, и handler молчит: он привязан к фактическому изменению,
а не к запуску playbook.
