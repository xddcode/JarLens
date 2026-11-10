import hashlib
import pathlib
import tarfile
import tempfile
import unittest
import zipfile

import sys
sys.path.insert(0, str(pathlib.Path(__file__).parent))

from runtime import adoptium_architecture, bundle, select_jre, sha256


class RuntimeTests(unittest.TestCase):
    def test_adoptium_architecture(self):
        self.assertEqual(adoptium_architecture("amd64"), "x64")
        self.assertEqual(adoptium_architecture("arm64"), "aarch64")
        with self.assertRaises(ValueError):
            adoptium_architecture("x64")

    def test_select_jre(self):
        package = {"os": "windows", "architecture": "x64"}
        assets = [{"binary": {"image_type": "jre", "os": "windows", "architecture": "x64", "package": package}}]
        self.assertEqual(select_jre(assets, "windows", "x64"), package)
        with self.assertRaises(ValueError):
            select_jre(assets, "linux", "aarch64")

    def test_bundle_zip_and_tar(self):
        for target, suffix in (("windows", ".zip"), ("linux", ".tar.gz")):
            with self.subTest(target=target), tempfile.TemporaryDirectory() as temporary:
                root = pathlib.Path(temporary)
                runtime = root / "runtime"
                java = runtime / "bin" / "java"
                java.parent.mkdir(parents=True)
                java.write_bytes(b"runtime")
                archive = root / f"jarlens{suffix}"
                if suffix == ".zip":
                    with zipfile.ZipFile(archive, "w") as output:
                        output.writestr("jarlens.exe", b"application")
                else:
                    with tarfile.open(archive, "w:gz") as output:
                        output.add(java, arcname="jarlens")
                bundle(archive, runtime, target)
                if suffix == ".zip":
                    with zipfile.ZipFile(archive) as result:
                        self.assertEqual(result.read("jre/bin/java"), b"runtime")
                        self.assertEqual(result.read("jarlens.exe"), b"application")
                else:
                    with tarfile.open(archive, "r:gz") as result:
                        self.assertEqual(result.extractfile("jre/bin/java").read(), b"runtime")

    def test_sha256(self):
        with tempfile.TemporaryDirectory() as temporary:
            path = pathlib.Path(temporary) / "data"
            path.write_bytes(b"jarlens")
            self.assertEqual(sha256(path), hashlib.sha256(b"jarlens").hexdigest())


if __name__ == "__main__":
    unittest.main()
