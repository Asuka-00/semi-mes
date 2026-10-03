export interface MesOption {
  label: string;
  value: string | number;
}

export interface MesRelation {
  resource: string;
  listKey: string;
  labelKey: string;
  extraLabelKey?: string;
}

export interface MesField {
  key: string;
  label: App.I18n.I18nKey;
  type?: 'text' | 'number' | 'textarea' | 'status' | 'yesno' | 'select' | 'password';
  search?: boolean;
  searchColumn?: string;
  required?: boolean;
  options?: MesOption[];
  relation?: MesRelation;
  table?: boolean;
  form?: boolean;
}

export interface MesChildren {
  resource: string;
  listKey: string;
  perm: string;
  parentKey: string;
  parentColumn: string;
  sequenceKey: string;
  title?: App.I18n.I18nKey;
  fields: MesField[];
}

export type MesAssignment = 'user-roles' | 'role-menus';
