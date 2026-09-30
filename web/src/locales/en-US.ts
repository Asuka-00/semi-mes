export default {
  common: {
    add: 'Add',
    edit: 'Edit',
    delete: 'Delete',
    query: 'Query',
    reset: 'Reset',
    submit: 'Submit',
    cancel: 'Cancel',
    confirm: 'Confirm',
    save: 'Save',
    search: 'Search',
    action: 'Action',
    status: 'Status',
    enable: 'Enable',
    disable: 'Disable',
    view: 'View',
    export: 'Export',
    import: 'Import',
    refresh: 'Refresh',
    back: 'Back',
    success: 'Success',
    failed: 'Failed',
    loading: 'Loading...',
    noData: 'No Data',
    total: 'Total {total} items',
    confirmDelete: 'Are you sure to delete?',
    yes: 'Yes',
    no: 'No'
  },
  menu: {
    dashboard: 'Dashboard',
    system: 'System',
    'system.user': 'User',
    'system.role': 'Role',
    'system.menu': 'Menu',
    baseData: 'Base Data',
    'baseData.factory': 'Factory',
    'baseData.workshop': 'Workshop',
    'baseData.line': 'Production Line',
    'baseData.product': 'Product',
    'baseData.route': 'Process Route',
    'baseData.operation': 'Operation',
    'baseData.recipe': 'Recipe',
    workOrder: 'Work Order',
    'workOrder.list': 'Work Order List',
    lot: 'Lot',
    'lot.list': 'Lot List',
    wip: 'WIP',
    'wip.move': 'Move History',
    equipment: 'Equipment',
    'equipment.list': 'Equipment List',
    quality: 'Quality',
    'quality.inspection': 'Inspection'
  },
  button: {
    query: 'Query',
    add: 'Add',
    edit: 'Edit',
    delete: 'Delete'
  },
  auth: {
    login: 'Login',
    logout: 'Logout',
    username: 'Username',
    password: 'Password',
    rememberMe: 'Remember Me',
    forgotPassword: 'Forgot Password',
    welcomeBack: 'Welcome Back',
    pleaseLogin: 'Please login to your account'
  },
  user: {
    username: 'Username',
    realName: 'Real Name',
    email: 'Email',
    phone: 'Phone',
    status: 'Status',
    roles: 'Roles',
    createTime: 'Create Time',
    updateTime: 'Update Time'
  },
  baseData: {
    factory: {
      title: 'Factory Management',
      code: 'Factory Code',
      name: 'Factory Name',
      address: 'Address',
      contact: 'Contact',
      phone: 'Phone'
    },
    workshop: {
      title: 'Workshop Management',
      code: 'Workshop Code',
      name: 'Workshop Name',
      type: 'Workshop Type',
      factory: 'Factory',
      description: 'Description'
    },
    line: {
      title: 'Production Line Management',
      code: 'Line Code',
      name: 'Line Name',
      workshop: 'Workshop',
      capacity: 'Capacity',
      description: 'Description'
    },
    product: {
      title: 'Product Management',
      code: 'Product Code',
      name: 'Product Name',
      type: 'Product Type',
      version: 'Version',
      description: 'Description'
    },
    route: {
      title: 'Process Route Management',
      code: 'Route Code',
      name: 'Route Name',
      product: 'Product',
      version: 'Version',
      isDefault: 'Default Route',
      description: 'Description'
    },
    operation: {
      title: 'Operation Management',
      code: 'Operation Code',
      name: 'Operation Name',
      route: 'Route',
      type: 'Operation Type',
      sequence: 'Sequence',
      standardTime: 'Standard Time (Minutes)',
      description: 'Description'
    },
    recipe: {
      title: 'Recipe Management',
      code: 'Recipe Code',
      name: 'Recipe Name',
      operation: 'Operation',
      version: 'Version',
      parameters: 'Parameters',
      isDefault: 'Default Recipe',
      description: 'Description'
    }
  }
}
