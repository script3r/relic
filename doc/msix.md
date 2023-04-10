# Signing MSIX packages

## Context

MSIX is a Windows app package format that was introduced in Windows 10. It is an evolution of the previous APPX format, and it is designed to make application installation, distribution, and updating more efficient and secure. MSIX packages are container files that can contain all the necessary resources, dependencies, and metadata needed for an application to run on a Windows system. They use a combination of technologies such as containerization, digital signatures, and runtime APIs to ensure that the apps are secure, reliable, and easy to manage. MSIX packages can be installed through various methods, including the Microsoft Store, Microsoft Endpoint Manager, and standalone installation packages.

## Requirements

You should have a `relic` configuration file, you can find more details by visiting the [relevant documentation](https://github.com/sassoftware/relic/blob/6e27811296dcc5beef771a67d09880f1a1da07e2/doc/relic.yml). 

If you're looking for a minimal proof of concept, create the following in `~/.config/relic:`

```yml
---
tokens:
  file:
    type: file
keys:
  my_file_key:
    token: file
    keyfile: /path/to/your/key/rsa2048.key
    x509certificate: /path/to/your/cert/rsa2048.crt
```

For PKCS11 backed keys, see documentation. If you don't have sample certificates and keys, for testing purposes you can use the ones in this [repo](functest/testkeys). 

## Signing

To sign a single MSIX package, use the following:

```bash
./relic sign --file /path/to/package.msix --key my_file_key
```

Currently, to sign an MSIX bundle, you must first sign each individual MSIX package, use a tool to package the final file, and then sign the final file.

