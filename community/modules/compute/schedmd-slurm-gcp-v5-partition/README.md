## Description

This module creates a compute partition that can be used as input to the
[schedmd-slurm-gcp-v5-controller](../../scheduler/schedmd-slurm-gcp-v5-controller/README.md).

The partition module is designed to work alongside the
[schedmd-slurm-gcp-v5-node-group](../schedmd-slurm-gcp-v5-node-group/README.md)
module. A partition can be made up of one or
more node groups, provided either through `use` (preferred) or defined manually
in the `node_groups` variable.

> **Warning**: updating a partition and running `terraform apply` will not cause
> the slurm controller to update its own configurations (`slurm.conf`) unless
> `enable_reconfigure` is set to true in the partition and controller modules.

### Example

The following code snippet creates a partition module with:

* 2 node groups added via `use`.
  * The first node group is made up of machines of type `c2-standard-30`.
  * The second node group is made up of machines of type `c2-standard-60`.
  * Both node groups have a maximum count of 200 dynamically created nodes.
* partition name of "compute".
* connected to the `network1` module via `use`.
* nodes mounted to homefs via `use`.

```yaml
- id: node_group_1
  source: community/modules/compute/schedmd-slurm-gcp-v5-node-group
  settings:
    name: c30
    node_count_dynamic_max: 200
    machine_type: c2-standard-30

- id: node_group_2
  source: community/modules/compute/schedmd-slurm-gcp-v5-node-group
  settings:
    name: c60
    node_count_dynamic_max: 200
    machine_type: c2-standard-60

- id: compute_partition
  source: community/modules/compute/schedmd-slurm-gcp-v5-partition
  use:
  - network1
  - homefs
  - node_group_1
  - node_group_2
  settings:
    partition_name: compute
```

For a complete example using this module, see
[slurm-gcp-v5-cluster.yaml](../../../examples/slurm-gcp-v5-cluster.yaml).

### Compute VM Zone Policies

> **_WARNING:_** Lenient zone policies can lead to additional egress costs when
> moving data between Google Cloud resources in different zones in the same
> region, such as between filestore and other VM instances. For more information
> on egress fees, see the [Network Pricing][networkpricing] Google Cloud
> documentation.
>
> To avoid egress charges, ensure your compute nodes are created in the same
> zone as the other resources that share data with them by setting
> `zone_policy_deny` to all other zones in the region.

The Slurm on GCP partition modules provide the option to set policies regarding
which zone the compute VM instances will be created in through the
`zone_policy_allow` and `zone_policy_deny` variables.

As an example, see the the following module:

```yaml
- id: partition-with-zone-policy
  source: community/modules/compute/schedmd-slurm-gcp-v5-partition
  settings:
    zone_policy_allow:
    - us-central1-a
    - us-central1-b
    zone_policy_deny: [us-central1-f]
```

In this module, the following is defined:

* `us-central1-a` and `us-central1-b` zones have been explicitly allowed.
* `us-central1-f` has been explicitly denied, therefore no nodes in this
  partition will be created in that zone.
* Since `us-central1-c` was not included in the zone policy, it will default to
  "Allow", which means the partition has the same likelihood of creating a node in
  that zone as the zones explicitly listed under `zone_policy_allow`.

> **_NOTE:_** `zone_policy_allow` does not guarantee the use of specified zones
> because zones are allowed by default. Configure `zone_policy_deny` to ensure
> that zones outside the allowed list are not used.

#### Setting a Single Zone

The `zone` variable is another option for setting the zone policy. If `zone` is
set and neither `zone_policy_deny` nor `zone_policy_allow` are set, the
policy will be configured as follows:

* All _currently active_ zones in the region **at deploy time** will be set in the
 `zone_policy_deny` list, with the exception of the provided `zone`.
* The provided `zone` will be set as the only value in the `zone_policy_allow`
  list.

`zone_policy_allow` and `zone_policy_deny` take precedence over `zone` if both
are set.

> **_NOTE:_** If a new zone is added to the region while the cluster is active,
> nodes in the partition may be created in that zone as well. In this case, the
> partition may need to be redeployed (possible via `enable_reconfigure` if set)
> to ensure the newly added zone is set to "Deny".

[networkpricing]: https://cloud.google.com/vpc/network-pricing

## Support
The HPC Toolkit team maintains the wrapper around the [slurm-on-gcp] terraform
modules. For support with the underlying modules, see the instructions in the
[slurm-gcp README][slurm-gcp-readme].

[slurm-on-gcp]: https://github.com/SchedMD/slurm-gcp
[slurm-gcp-readme]: https://github.com/SchedMD/slurm-gcp#slurm-on-google-cloud-platform

## License
<!-- BEGINNING OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
Copyright 2022 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 0.13.0 |
| <a name="requirement_google"></a> [google](#requirement\_google) | >= 3.83 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_google"></a> [google](#provider\_google) | >= 3.83 |

## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_slurm_partition"></a> [slurm\_partition](#module\_slurm\_partition) | github.com/SchedMD/slurm-gcp.git//terraform/slurm_cluster/modules/slurm_partition | 5.6.2 |

## Resources

| Name | Type |
|------|------|
| [google_compute_zones.available](https://registry.terraform.io/providers/hashicorp/google/latest/docs/data-sources/compute_zones) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_deployment_name"></a> [deployment\_name](#input\_deployment\_name) | Name of the deployment. | `string` | n/a | yes |
| <a name="input_enable_placement"></a> [enable\_placement](#input\_enable\_placement) | Enable placement groups. | `bool` | `true` | no |
| <a name="input_enable_reconfigure"></a> [enable\_reconfigure](#input\_enable\_reconfigure) | Enables automatic Slurm reconfigure on when Slurm configuration changes (e.g.<br>slurm.conf.tpl, partition details). Compute instances and resource policies<br>(e.g. placement groups) will be destroyed to align with new configuration.<br><br>NOTE: Requires Python and Google Pub/Sub API.<br><br>*WARNING*: Toggling this will impact the running workload. Deployed compute nodes<br>will be destroyed and their jobs will be requeued. | `bool` | `false` | no |
| <a name="input_exclusive"></a> [exclusive](#input\_exclusive) | Exclusive job access to nodes. | `bool` | `true` | no |
| <a name="input_is_default"></a> [is\_default](#input\_is\_default) | Sets this partition as the default partition by updating the partition\_conf.<br>If "Default" is already set in partition\_conf, this variable will have no effect. | `bool` | `false` | no |
| <a name="input_network_storage"></a> [network\_storage](#input\_network\_storage) | An array of network attached storage mounts to be configured on the partition compute nodes. | <pre>list(object({<br>    server_ip             = string,<br>    remote_mount          = string,<br>    local_mount           = string,<br>    fs_type               = string,<br>    mount_options         = string,<br>    client_install_runner = map(string)<br>    mount_runner          = map(string)<br>  }))</pre> | `[]` | no |
| <a name="input_node_groups"></a> [node\_groups](#input\_node\_groups) | A list of node groups associated with this partition. See<br>schedmd-slurm-gcp-v5-node-group for more information on defining a node<br>group in a blueprint. | <pre>list(object({<br>    access_config = list(object({<br>      network_tier = string<br>    }))<br>    node_count_static      = number<br>    node_count_dynamic_max = number<br>    group_name             = string<br>    node_conf              = map(string)<br>    additional_disks = list(object({<br>      disk_name    = string<br>      device_name  = string<br>      disk_size_gb = number<br>      disk_type    = string<br>      disk_labels  = map(string)<br>      auto_delete  = bool<br>      boot         = bool<br>    }))<br>    bandwidth_tier         = string<br>    can_ip_forward         = bool<br>    disable_smt            = bool<br>    disk_auto_delete       = bool<br>    disk_labels            = map(string)<br>    disk_size_gb           = number<br>    disk_type              = string<br>    enable_confidential_vm = bool<br>    enable_oslogin         = bool<br>    enable_shielded_vm     = bool<br>    enable_spot_vm         = bool<br>    gpu = object({<br>      count = number<br>      type  = string<br>    })<br>    instance_template   = string<br>    labels              = map(string)<br>    machine_type        = string<br>    metadata            = map(string)<br>    min_cpu_platform    = string<br>    on_host_maintenance = string<br>    preemptible         = bool<br>    service_account = object({<br>      email  = string<br>      scopes = list(string)<br>    })<br>    shielded_instance_config = object({<br>      enable_integrity_monitoring = bool<br>      enable_secure_boot          = bool<br>      enable_vtpm                 = bool<br>    })<br>    spot_instance_config = object({<br>      termination_action = string<br>    })<br>    source_image_family  = string<br>    source_image_project = string<br>    source_image         = string<br>    tags                 = list(string)<br>  }))</pre> | `[]` | no |
| <a name="input_partition_conf"></a> [partition\_conf](#input\_partition\_conf) | Slurm partition configuration as a map.<br>See https://slurm.schedmd.com/slurm.conf.html#SECTION_PARTITION-CONFIGURATION | `map(string)` | `{}` | no |
| <a name="input_partition_name"></a> [partition\_name](#input\_partition\_name) | The name of the slurm partition. | `string` | n/a | yes |
| <a name="input_partition_startup_scripts_timeout"></a> [partition\_startup\_scripts\_timeout](#input\_partition\_startup\_scripts\_timeout) | The timeout (seconds) applied to the partition startup script. If<br>any script exceeds this timeout, then the instance setup process is considered<br>failed and handled accordingly.<br><br>NOTE: When set to 0, the timeout is considered infinite and thus disabled. | `number` | `300` | no |
| <a name="input_project_id"></a> [project\_id](#input\_project\_id) | Project in which the HPC deployment will be created. | `string` | n/a | yes |
| <a name="input_region"></a> [region](#input\_region) | The default region for Cloud resources. | `string` | n/a | yes |
| <a name="input_slurm_cluster_name"></a> [slurm\_cluster\_name](#input\_slurm\_cluster\_name) | Cluster name, used for resource naming and slurm accounting. If not provided it will default to the first 8 characters of the deployment name (removing any invalid characters). | `string` | `null` | no |
| <a name="input_startup_script"></a> [startup\_script](#input\_startup\_script) | Startup script that will be used by the partition VMs. | `string` | `""` | no |
| <a name="input_subnetwork_project"></a> [subnetwork\_project](#input\_subnetwork\_project) | The project the subnetwork belongs to. | `string` | `""` | no |
| <a name="input_subnetwork_self_link"></a> [subnetwork\_self\_link](#input\_subnetwork\_self\_link) | Subnet to deploy to. | `string` | `null` | no |
| <a name="input_zone"></a> [zone](#input\_zone) | Zone in which to create all compute VMs. If `zone_policy_deny` or `zone_policy_allow` are set, the `zone` variable will be ignored. | `string` | `null` | no |
| <a name="input_zone_policy_allow"></a> [zone\_policy\_allow](#input\_zone\_policy\_allow) | Partition nodes will prefer to be created in the listed zones. If a zone appears<br>in both zone\_policy\_allow and zone\_policy\_deny, then zone\_policy\_deny will take<br>priority for that zone. | `set(string)` | `[]` | no |
| <a name="input_zone_policy_deny"></a> [zone\_policy\_deny](#input\_zone\_policy\_deny) | Partition nodes will not be created in the listed zones. If a zone appears in<br>both zone\_policy\_allow and zone\_policy\_deny, then zone\_policy\_deny will take<br>priority for that zone. | `set(string)` | `[]` | no |
| <a name="input_zone_target_shape"></a> [zone\_target\_shape](#input\_zone\_target\_shape) | Strategy for distributing VMs across zones in a region.<br>ANY<br>  GCE picks zones for creating VM instances to fulfill the requested number of VMs<br>  within present resource constraints and to maximize utilization of unused zonal<br>  reservations.<br>ANY\_SINGLE\_ZONE (default)<br>  GCE always selects a single zone for all the VMs, optimizing for resource quotas,<br>  available reservations and general capacity.<br>BALANCED<br>  GCE prioritizes acquisition of resources, scheduling VMs in zones where resources<br>  are available while distributing VMs as evenly as possible across allowed zones<br>  to minimize the impact of zonal failure. | `string` | `"ANY_SINGLE_ZONE"` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_partition"></a> [partition](#output\_partition) | Details of a slurm partition |
<!-- END OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
