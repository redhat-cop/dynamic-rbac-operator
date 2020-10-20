<br />
<p align="center">

  <h3 align="center">Dynamic RBAC Operator</h3>

  <p align="center">
    Flexible definitions of Kubernetes RBAC rules
  </p>
</p>

<!-- TABLE OF CONTENTS -->

## Table of Contents

- [About the Project](#about-the-project)
  - [Built With](#built-with)
- [Getting Started](#getting-started)
  - [Installation](#installation)
- [Usage](#usage)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [License](#license)
- [Contact](#contact)

<!-- ABOUT THE PROJECT -->

## About The Project

Writing Kubernetes RBAC definitions by hand can be a pain. This operator allows you to define "Dynamic" RBAC rules that change based on the state of your cluster, so you can spend your time writing the RBAC _patterns_ that you'd like to deploy, rather than traditional, fully enumerated RBAC rules.

### Built With

- [Operator-SDK](https://github.com/operator-framework/operator-sdk)

<!-- GETTING STARTED -->

## Getting Started

### Installation

This operator can be installed with Kustomize:

`kustomize build config/default | oc apply -f -`

<!-- USAGE EXAMPLES -->

## Usage

Once the operator is installed, you can begin using `DynamicRole` and `DynamicClusterRole` resources within your cluster.

For example, the `DynamicClusterRole`:

```yaml
apiVersion: rbac.redhatcop.redhat.io/v1alpha1
kind: DynamicClusterRole
metadata:
  name: admin-without-users
spec:
  inherit:
    - name: cluster-admin
      kind: ClusterRole
  deny:
    - apiGroups:
        - "user.openshift.io"
      resources:
        - "users"
      verbs:
        - "*"
```

will cause the operator to use the cluster's resource discovery API to enumerate all of the individual permissions of the `cluster-admin` user, and then remove access to `user.openshift.io/users` resources.

You can then create a `RoleBinding` or `ClusterRoleBinding` to `admin-without-users` (as a `ClusterRole`) as normal, and permissions will work as expected!

<!-- ROADMAP -->

## Roadmap

See the [open issues](https://github.com/redhat-cop/dynamic-rbac-operator/issues) for a list of proposed features.

## Known Issues

1. Only one role can be inherited right now, even though it is spec'd as a list, because ruleset merging is still WIP.
2. Allow lists are in the spec but not yet implemented, because of the same reason as above.
3. This operator requires `cluster-admin` privileges, because it needs to be able to write RBAC rules that grant arbitrary permissions that it doesn't actually need itself. `make manifests` currently overwrites this.

<!-- CONTRIBUTING -->

## Contributing

Contributions are what make the open source community such an amazing place to be learn, inspire, and create. Any contributions you make are **greatly appreciated**.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

<!-- LICENSE -->

## License

Distributed under the Apache License 2.0. See `LICENSE` for more information.

<!-- CONTACT -->

## Contact

Project Link: [https://github.com/redhat-cop/dynamic-rbac-operator](https://github.com/redhat-cop/dynamic-rbac-operator)
