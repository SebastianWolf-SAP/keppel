<!--
SPDX-FileCopyrightText: 2025 SAP SE

SPDX-License-Identifier: Apache-2.0
-->

# Auth driver: `keystone`

An auth driver using the Keystone V3 API of an OpenStack cluster. With this driver, Keppel auth tenants correspond to
Keystone projects.

- Requests to the [Keppel API](../api-spec.md) are authenticated by reading a Keystone token from the X-Auth-Token
  request header. When using the client commands of Keppel, the regular OpenStack auth environment variables (`OS_...`)
  need to be present. See [documentation for openstackclient][os-env] for details.
- Requests to the Docker Registry API can be authenticated with username and password, and the username has one of the
  following formats:
  ```
  user_name@user_domain_name/project_name@project_domain_name
  user_name@domain_name/project_name
  ```
  The latter format implies that user and project are located in the same domain.
- Requests to the Docker Registry API can also be authenticated with an application credential by giving the user name
  `applicationcredential-`, followed by the application credential ID. The supplied password must be the application
  credential secret. It's not yet possible to identify an application credential by its name, but a syntax for this
  could be added in a later release.

## Server-side configuration

| Variable | Default | Explanation |
| -------- | ------- | ----------- |
| `OS_...` | *(required)* | A full set of OpenStack auth environment variables for Keppel's service user. See [documentation for openstackclient][os-env] for details. |
| `KEPPEL_OSLO_POLICY_PATH` | *(required)* | Path to the `policy.[json|yaml]` file for this service. |

Keppel understands access rules in the [`oslo.policy` JSON][os-pol-json] and [`oslo.policy` YAML][os-pol-yaml] format. An example can be seen at
[`docs/example-policy.json`](../example-policy.json). The following rules are expected:

- `account:list` is required for any non-anonymous access to the API.
- `account:show` enables read access to repository and tag listings.
- `account:pull` allows to `docker pull` images.
- `account:push` allows to `docker push` images.
- `account:delete` allows to delete image manifests and tags.
- `account:edit` enables write access to an account's configuration.
- `quota:show` enables read access to a project's quotas and usage statistics.
- `quota:edit` enables write access to a project's quotas.

All policy rules can use the object attribute `%(target.project.id)s`.

### Keystone service catalog

- The top-level path of the Keppel API (e.g. `https://keppel.example.com/`) should be entered in the service catalog as service type `keppel`.
- If integration with [Limes][limes] is desired, the `/liquid/` subpath of the Keppel API (e.g. `https://keppel.example.com/liquid/`) can be entered in the service catalog as service type `liquid-keppel`.

See also: [List of available API attributes](https://github.com/sapcc/go-bits/blob/53eeb20fde03c3d0a35e76cf9c9a06b63a415e6b/gopherpolicy/pkg.go#L151-L164)

[limes]: https://github.com/sapcc/limes
[os-env]: https://docs.openstack.org/python-openstackclient/latest/cli/man/openstack.html
[os-pol-json]: https://docs.openstack.org/oslo.policy/latest/admin/policy-json-file.html
[os-pol-yaml]: https://docs.openstack.org/oslo.policy/latest/admin/policy-yaml-file.html
