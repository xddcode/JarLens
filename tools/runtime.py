#!/usr/bin/env python3
import argparse
import hashlib
import json
import os
import pathlib
import shutil
import tarfile
import tempfile
import urllib.request
import zipfile

API = "https://api.adoptium.net/v3/assets/latest/21/hotspot"


def download_json(url):
    request = urllib.request.Request(url, headers={"User-Agent": "JarLens/0.1"})
    with urllib.request.urlopen(request, timeout=60) as response:
        return json.load(response)


def download_file(url, destination):
    request = urllib.request.Request(url, headers={"User-Agent": "JarLens/0.1"})
    with urllib.request.urlopen(request, timeout=300) as response, open(destination, "wb") as output:
        shutil.copyfileobj(response, output)


def sha256(path):
    digest = hashlib.sha256()
    with open(path, "rb") as source:
        for chunk in iter(lambda: source.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def select_jre(assets, os_name, architecture):
    for asset in assets:
        binary = asset.get("binary", {})
        if binary.get("image_type") != "jre":
            continue
        package = binary.get("package", {})
        if binary.get("os") == os_name and binary.get("architecture") == architecture:
            return package
    raise ValueError(f"no JRE for {os_name}/{architecture}")


def adoptium_architecture(architecture):
    if architecture == "arm64":
        return "aarch64"
    if architecture == "amd64":
        return "x64"
    raise ValueError(f"unsupported architecture: {architecture}")


def download_jre(os_name, architecture, output):
    os_name = "mac" if os_name == "darwin" else os_name
    architecture = adoptium_architecture(architecture)
    query = f"?os={os_name}&architecture={architecture}&image_type=jre&vendor=eclipse"
    assets = download_json(API + query)
    package = select_jre(assets, os_name, architecture)
    output = pathlib.Path(output)
    output.mkdir(parents=True, exist_ok=True)
    with tempfile.TemporaryDirectory(prefix="jarlens-jre-") as temporary:
        archive = pathlib.Path(temporary) / package["name"]
        download_file(package["link"], archive)
        if sha256(archive) != package["checksum"].lower():
            raise ValueError("JRE checksum mismatch")
        extracted = pathlib.Path(temporary) / "jre"
        extracted.mkdir()
        if archive.name.endswith(".zip"):
            with zipfile.ZipFile(archive) as source:
                source.extractall(extracted)
        else:
            with tarfile.open(archive, "r:gz") as source:
                source.extractall(extracted)
        children = list(extracted.iterdir())
        if len(children) == 1 and children[0].is_dir():
            extracted = children[0]
        target = output / "jre"
        if target.exists():
            shutil.rmtree(target)
        shutil.copytree(extracted, target)


def bundle(archive_path, runtime, target):
    archive_path = pathlib.Path(archive_path)
    runtime = pathlib.Path(runtime)
    if not runtime.is_dir():
        raise ValueError(f"runtime directory not found: {runtime}")
    if archive_path.suffix == ".zip":
        bundle_zip(archive_path, runtime, target)
    elif archive_path.name.endswith(".tar.gz"):
        bundle_tar(archive_path, runtime)
    else:
        raise ValueError(f"unsupported archive: {archive_path}")


def bundle_zip(archive_path, runtime, target):
    prefix = "JarLens.app/Contents/Resources/jre" if target == "darwin" else "jre"
    temporary = archive_path.with_suffix(archive_path.suffix + ".tmp")
    with zipfile.ZipFile(archive_path, "r") as source, zipfile.ZipFile(temporary, "w", zipfile.ZIP_DEFLATED) as destination:
        for item in source.infolist():
            destination.writestr(item, source.read(item.filename))
        for path in sorted(runtime.rglob("*")):
            if path.is_file():
                destination.write(path, f"{prefix}/{path.relative_to(runtime).as_posix()}")
    os.replace(temporary, archive_path)


def bundle_tar(archive_path, runtime):
    temporary = archive_path.with_suffix(".tmp")
    with tarfile.open(archive_path, "r:gz") as source, tarfile.open(temporary, "w:gz") as destination:
        for item in source.getmembers():
            destination.addfile(item, source.extractfile(item))
        for path in sorted(runtime.rglob("*")):
            if path.is_file():
                destination.add(path, arcname=f"jre/{path.relative_to(runtime).as_posix()}")
    os.replace(temporary, archive_path)


def main():
    parser = argparse.ArgumentParser(description="Download and bundle a pinned Adoptium JRE")
    subparsers = parser.add_subparsers(dest="command", required=True)
    download = subparsers.add_parser("download")
    download.add_argument("-os", required=True)
    download.add_argument("-arch", required=True)
    download.add_argument("-out", required=True)
    bundle_parser = subparsers.add_parser("bundle")
    bundle_parser.add_argument("-archive", required=True)
    bundle_parser.add_argument("-runtime", required=True)
    bundle_parser.add_argument("-target", required=True)
    args = parser.parse_args()
    if args.command == "download":
        download_jre(args.os, args.arch, args.out)
    else:
        bundle(args.archive, args.runtime, args.target)


if __name__ == "__main__":
    main()
