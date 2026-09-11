#!/usr/bin/env python3
"""Render the real inventory.tf with Terraform, without a cloud provider or state.

Only a temporary, provider-free fixture is planned. Real variables.tf and the
VM module's input types are retained; its output IPs come from test fixtures.
"""
import argparse
import copy
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import yaml


ROOT = Path(__file__).resolve().parents[1]
NAMES = ("lb1", "lb2", "node1", "node2", "node3")
INVENTORIES = ("kubespray_inventory", "kubespray_all_yml", "external_ha_inventory")
GROUP_ROLES = {"kube_control_plane": "control-plane", "etcd": "etcd", "kube_node": "worker"}
ENV = dict(os.environ, CHECKPOINT_DISABLE="1", TF_IN_AUTOMATION="1")
# Do not inherit user arguments, state location, or cloud authentication.
for key in list(ENV):
    if key.startswith(("TF_VAR_", "TF_CLI_ARGS", "YC_")) or key in ("TF_DATA_DIR", "TF_WORKSPACE"):
        ENV.pop(key)


def run(args, cwd, *, env=ENV, ok=True):
    result = subprocess.run(args, cwd=cwd, env=env, capture_output=True, text=True)
    if ok and result.returncode:
        raise AssertionError(f"{' '.join(map(str, args))}\n{result.stdout}\n{result.stderr}")
    return result


def write_json(path, value):
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2) + "\n")


def addresses(names):
    return {name: {"external_ip": f"192.0.2.{i}", "internal_ip": f"10.10.0.{i}"}
            for i, name in enumerate(sorted(names), 11)}


def parse_inventory(directory, filename, text):
    path = directory / filename
    path.write_text(text)
    env = dict(ENV, ANSIBLE_CONFIG=str(directory / "ansible.cfg"),
               ANSIBLE_LOCAL_TEMP=str(directory / "ansible-tmp"),
               ANSIBLE_INVENTORY_ENABLED="ini")
    return json.loads(run(["ansible-inventory", "-i", str(path), "--list"], directory, env=env).stdout)


def render(source=ROOT, topology=None, ips=None, expected_error=None):
    with tempfile.TemporaryDirectory(prefix="video2a-inventory-test-") as tmp:
        directory = Path(tmp)
        for name in ("locals.tf", "inventory.tf", "outputs.tf"):
            shutil.copyfile(source / name, directory / name)
        shutil.copyfile(ROOT / "variables.tf", directory / "variables.tf")
        fake = directory / "fake-vm"
        fake.mkdir()
        # This is the real module input schema: extra inventory metadata must
        # not change the roles/resources received by the VM resource.
        shutil.copyfile(ROOT / "modules/linux-vm/variables.tf", fake / "variables.tf")
        write_json(fake / "main.tf.json", {
            "variable": {"fixture_addresses": {"type": "any"}},
            "output": {
                "nodes": {"value": "${{ for name, node in var.nodes : name => { external_ip = var.fixture_addresses[name].external_ip, internal_ip = var.fixture_addresses[name].internal_ip, roles = node.roles, instance_id = name } }}"},
                "typed_nodes": {"value": "${var.nodes}"},
            },
        })
        write_json(directory / "fixture.tf.json", {
            "variable": {"fixture_addresses": {"type": "any"}},
            "module": {"linux_vm": {
                "source": "./fake-vm", "nodes": "${local.nodes}",
                "cluster_name": "${var.cluster_name}", "zone": "${var.zone}",
                "image_id": "fixture", "subnet_id": "fixture", "security_group_ids": [],
                "ssh_user": "${var.ssh_user}", "ssh_public_key": "fixture-public-key",
                "fixture_addresses": "${var.fixture_addresses}",
            }},
            "output": {
                "test_topology": {"value": "${local.nodes}"},
                "test_typed_nodes": {"value": "${module.linux_vm.typed_nodes}"},
            },
        })
        if topology is not None:
            write_json(directory / "nodes_override.tf.json", {"locals": {"nodes": topology}})
        write_json(directory / "fixture.auto.tfvars.json", {
            "allowed_ssh_cidr": "192.0.2.1/32",
            "fixture_addresses": ips or addresses(topology or NAMES),
        })
        (directory / "ansible.cfg").write_text("[defaults]\n")
        run(["terraform", "init", "-backend=false", "-input=false", "-no-color"], directory)
        planned = run(["terraform", "plan", "-input=false", "-refresh=false", "-lock=false",
                       "-no-color", "-out=fixture.tfplan"], directory, ok=expected_error is None)
        if expected_error is not None:
            assert planned.returncode != 0, "Invalid topology was accepted"
            assert expected_error in planned.stdout + planned.stderr
            return None
        plan = json.loads(run(["terraform", "show", "-json", "fixture.tfplan"], directory).stdout)
        assert not plan.get("resource_changes"), "The fixture must never contain cloud resources"
        outputs = {key: item["value"] for key, item in plan["planned_values"]["outputs"].items()}
        outputs["parsed_kubespray"] = parse_inventory(directory, "kubespray.ini", outputs["kubespray_inventory"])
        outputs["parsed_ha"] = parse_inventory(directory, "external.ini", outputs["external_ha_inventory"])
        return outputs


def check_groups(result):
    topology = result["test_topology"]
    kube = result["parsed_kubespray"]
    ha = result["parsed_ha"]
    kube_names = sorted(name for name, node in topology.items() if "load-balancer" not in node["roles"])
    lb_names = sorted(name for name, node in topology.items() if "load-balancer" in node["roles"])
    assert sorted(kube["_meta"]["hostvars"]) == kube_names
    for group, role in GROUP_ROLES.items():
        expected = sorted(name for name, node in topology.items() if role in node["roles"])
        assert sorted(kube.get(group, {}).get("hosts", [])) == expected, group
    assert sorted(kube["k8s_cluster"]["children"]) == ["kube_control_plane", "kube_node"]
    assert sorted(ha["kubernetes_nodes"]["hosts"]) == kube_names
    assert sorted(ha["load_balancers"]["hosts"]) == lb_names
    for name in lb_names:
        actual = ha["_meta"]["hostvars"][name]
        assert actual["ha_overlay_ip"] == topology[name]["ha"]["overlay_ip"]
        assert actual["ha_priority"] == topology[name]["ha"]["priority"]
    for name in kube_names:
        actual = kube["_meta"]["hostvars"][name]
        assert actual["ansible_host"] == result["nodes"][name]["external_ip"]
        assert actual["ip"] == actual["access_ip"] == result["nodes"][name]["internal_ip"]
    endpoint = [name for name, node in topology.items() if node.get("api_endpoint", False)]
    assert len(endpoint) == 1
    san_ips = [line.strip()[2:] for line in result["kubespray_all_yml"].splitlines() if line.strip().startswith("- ")]
    assert san_ips == [result["nodes"][endpoint[0]]["external_ip"]]


def equivalent(before, after):
    # Ansible ignores INI indentation/blank lines. Compare the actual consumer
    # output, not the old template's accidental whitespace around directives.
    assert before["parsed_kubespray"] == after["parsed_kubespray"]
    assert before["parsed_ha"] == after["parsed_ha"]
    assert yaml.safe_load(before["kubespray_all_yml"]) == yaml.safe_load(after["kubespray_all_yml"])


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--baseline-dir", type=Path)
    parser.add_argument("--state", type=Path, help="Read only the existing nodes/inventory outputs for equivalence")
    args = parser.parse_args()
    state_hash = hashlib.sha256(args.state.read_bytes()).hexdigest() if args.state else None
    original = render()
    check_groups(original)
    topology = original["test_topology"]
    print("PASS: current topology, IPs, API SAN and HA host settings")
    if args.baseline_dir:
        baseline = render(args.baseline_dir)
        equivalent(baseline, original)
        assert original["test_typed_nodes"] == baseline["test_typed_nodes"]
        print("PASS: inventories/SAN equivalent to baseline; VM inputs unchanged")
    if args.state:
        saved = json.loads(args.state.read_text())["outputs"]
        result = render(ips=saved["nodes"]["value"])
        with tempfile.TemporaryDirectory(prefix="video2a-saved-inventory-") as tmp:
            directory = Path(tmp)
            (directory / "ansible.cfg").write_text("[defaults]\n")
            stored = {
                "parsed_kubespray": parse_inventory(directory, "saved-kubespray.ini", saved["kubespray_inventory"]["value"]),
                "parsed_ha": parse_inventory(directory, "saved-ha.ini", saved["external_ha_inventory"]["value"]),
                "kubespray_all_yml": saved["kubespray_all_yml"]["value"],
            }
            equivalent(stored, result)
        print("PASS: inventories/SAN equivalent to existing state values")
    cases = {}
    added = copy.deepcopy(topology)
    added["node4"] = {"roles": ["worker"], "resources": dict(topology["node3"]["resources"])}
    cases["new worker"] = added
    changed = copy.deepcopy(topology)
    changed["node2"]["roles"].remove("worker")
    changed["node3"]["roles"].append("control-plane")
    cases["changed roles"] = changed
    cases["all machines renamed"] = {"renamed-" + name: node for name, node in topology.items()}
    removed = copy.deepcopy(topology)
    del removed["node3"]
    cases["worker removed"] = removed
    scaled = copy.deepcopy(topology)
    scaled["edge-new"] = {"roles": ["load-balancer"], "resources": dict(topology["lb1"]["resources"]),
                          "ha": {"overlay_ip": "10.77.0.13", "priority": 90}}
    cases["new load balancer"] = scaled
    for label, nodes in cases.items():
        check_groups(render(topology=nodes))
        print("PASS:", label)
    reordered = render(topology=dict(reversed(list(topology.items()))))
    for key in INVENTORIES:
        assert reordered[key] == original[key], f"Map order changed output: {key}"
    print("PASS: map order does not change output or HA host settings")
    for label in ("no endpoint", "two endpoints", "worker endpoint"):
        invalid = copy.deepcopy(topology)
        if label != "two endpoints":
            invalid["node1"].pop("api_endpoint")
        if label == "two endpoints":
            invalid["node2"]["api_endpoint"] = True
        if label == "worker endpoint":
            invalid["node3"]["api_endpoint"] = True
        render(topology=invalid, expected_error="ровно у одной машины")
        print("PASS: reject", label)
    if args.state:
        assert hashlib.sha256(args.state.read_bytes()).hexdigest() == state_hash
        print("PASS: real state unchanged")


if __name__ == "__main__":
    main()
