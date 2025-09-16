# Terraform Provider - AWS Extras

This is an AWS **Extras** provider. It provides various extras that the default/primary provider does not.

## Resources

- awsextras_terminate_instances
- awsextras_remove_key_pairs
- awsextras_secretsmanager_secret_rotation

### awsextras_terminate_instances

**Dangerous if not configured properly**

This allows you to identify aws instances in a subnet that are not managed terraform state and remove them on a destroy.

This will current filter on subnet ids and will try and identify instances that should be removed, you and provide a list
of instance ids that should be excluded automatically, otherwise it tries to determine if an instance should be removed
by its lack of tagging.

### awsextras_remove_key_pairs

### awsextras_secretsmanager_secret_rotation

This resource allows you to modify an existing secretId rotation details. The default AWS resource provider is setup to
require you to provide a lambda function and create a new rotation config, this provides a different set of options.