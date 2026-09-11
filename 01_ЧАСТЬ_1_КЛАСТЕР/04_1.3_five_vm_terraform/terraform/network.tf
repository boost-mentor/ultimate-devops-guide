# Root module связывает готовые блоки. Детали VPC/подсети/SG находятся в
# cohesive module: у него один входной контракт и явные outputs.
module "network" {
  source = "./modules/network"

  name_prefix         = var.cluster_name
  zone                = var.zone
  vpc_cidr            = var.vpc_cidr
  trusted_admin_cidrs = local.trusted_admin_cidrs
}
