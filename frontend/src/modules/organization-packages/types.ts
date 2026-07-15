export type OrganizationPackage = {
  id: number;
  package_name: string;
  package_description: string;
  technical_coefficient: number;
  technical_professional_coefficient: number;
  consumption_coefficient: number;
  subsidy_percentage: number;
  supplementary_percentage: number;
  organization_percentage: number;
};

export type OrganizationPackagePayload = {
  package_name: string;
  package_description: string;
  technical_coefficient: number;
  technical_professional_coefficient: number;
  consumption_coefficient: number;
  subsidy_percentage: number;
  supplementary_percentage: number;
  organization_percentage: number;
};
