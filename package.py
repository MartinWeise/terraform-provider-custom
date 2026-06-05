import json
import os
import shutil
import subprocess
import sys
import yaml


# Script that packages/bundles a valid Terraform registry
# @dev Only considers the latest version (lacks memory)
# @author Martin Weise <martin.weise@sequello.com>

def generate_metadata(host_url: str, version: str, gpg_fingerprint: str, goos: str, goarch: str,
                      out_path: str = "./registry", build_path: str = "./dist") -> None:
    for path in [f"{out_path}/.well-known", f"{out_path}/terraform-provider-custom/{version}"]:
        if not os.path.exists(path):
            os.makedirs(path)
    print(f"Creating registry directory skeleton... OK")

    path = f"{out_path}/.well-known/terraform.json"
    print(f"Creating registry discovery metadata... OK")
    data = {
        "providers.v1": "/v1/providers/"
    }
    with open(path, "w") as f:
        json.dump(data, f)

    files = [file for file in os.listdir(build_path) if os.path.isfile(os.path.join(build_path, file)) and
             file.startswith("custom")]
    for file in files:
        shutil.copy(f"{build_path}/{file}", f"{out_path}/terraform-provider-custom/{version}/{file}")
    print(f"Copying {len(files)} artifacts... OK")

    dir_path = f"{out_path}/v1/providers/sequello/custom/{version}/download/{goos}"
    if not os.path.exists(dir_path):
        os.makedirs(dir_path)

    path = f"{dir_path}/{goarch}"
    data = {
        "protocols": ["5.0"],
        "os": goos,
        "arch": goarch,
        "filename": f"custom_{version}_{goos}_{goarch}.zip",
        "download_url": f"{host_url}/terraform-provider-custom/{version}/terraform-provider-custom_{version}_{goos}_{goarch}.zip",
        "shasums_url": f"{host_url}/terraform-provider-custom/{version}/terraform-provider-custom_{version}_SHA256SUMS",
        "shasums_signature_url": f"{host_url}/terraform-provider-custom/{version}/terraform-provider-custom_{version}_SHA256SUMS.sig",
        "shasum": get_checksum(build_path, goos, goarch),
        "signing_keys": {
            "gpg_public_keys": [
                {
                    "key_id": gpg_fingerprint,
                    "ascii_armor": get_gpg_armor_str(gpg_fingerprint),
                    "trust_signature": "",
                    "source": "Sequello",
                    "source_url": "https://www.sequello.com/security.html"
                }
            ]
        }
    }
    with open(path, "w") as f:
        json.dump(data, f)
    print(f"Creating download artifact metadata... OK")


def matches(obj, goos, goarch):
    if (obj["type"] != "Archive" or "goos" not in obj or "goarch" not in obj):
        return False
    return obj["goos"] == goos and obj["goarch"] == goarch


def get_checksum(build_path: str, goos: str, goarch: str) -> str:
    with open(f"{build_path}/artifacts.json", "r") as f:
        data = json.load(f)
        filter = [item for item in data if matches(item, goos, goarch)]
        return filter[0]["extra"]["Checksum"].replace("sha256:", "")


def get_gpg_armor_str(gpg_fingerprint: str):
    print(f"Getting GPG armor of {gpg_fingerprint}... OK")
    result = subprocess.run(["gpg", "--armor", "--export", gpg_fingerprint], stdout=subprocess.PIPE)
    return result.stdout.decode("utf-8")


def load_binary_artifacts(build_path: str = "./dist") -> list:
    data = []
    with open(f"{build_path}/artifacts.json", "r") as f:
        data = json.load(f)
    return [b for b in data if b["type"] == "Binary"]


def generate_versions(version: str, artifacts: list, out_path: str = "./registry") -> None:
    versions = []
    for artifact in artifacts:
        versions.append(format_version(version, artifact["goos"], artifact["goarch"]))
    path = f"supported_old_versions.yaml"
    with open(path) as fs:
        try:
            contents = yaml.safe_load(fs)
            if contents:
                for key, metadata in contents.items():
                    if "artifacts" in metadata:
                        for artifact in metadata["artifacts"]:
                            versions.append(format_version(key, artifact["os"], artifact["arch"]))
        except yaml.YAMLError as exc:
            print(f"Error loading YAML {path}: {exc}")
            sys.exit(1)
    path = f"{out_path}/v1/providers/sequello/custom/versions"
    with open(path, "w") as f:
        json.dump({"versions": versions}, f)
    print(f"Versions: {set([v["version"] for v in versions])}")
    print(f"Created version list ... OK")


def format_version(version: str, os: str, arch: str) -> any:
    return {"version": version, "protocols": ["5.0"], "platforms": [{"os": os, "arch": arch}]}


if __name__ == "__main__":
    host_url = os.environ["HOST_URL"]
    version = os.environ["VERSION"]
    gpg_fingerprint = os.environ["GPG_KEY_ID"]
    artifacts = load_binary_artifacts()
    for artifact in artifacts:
        generate_metadata(host_url, version, gpg_fingerprint, artifact["goos"], artifact["goarch"])
    generate_versions(version, artifacts)
