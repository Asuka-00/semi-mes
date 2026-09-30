export default {
  common: {
    add: '新增',
    edit: '编辑',
    delete: '删除',
    query: '查询',
    reset: '重置',
    submit: '提交',
    cancel: '取消',
    confirm: '确认',
    save: '保存',
    search: '搜索',
    action: '操作',
    status: '状态',
    enable: '启用',
    disable: '禁用',
    view: '查看',
    export: '导出',
    import: '导入',
    refresh: '刷新',
    back: '返回',
    success: '成功',
    failed: '失败',
    loading: '加载中...',
    noData: '暂无数据',
    total: '共 {total} 条',
    confirmDelete: '确认删除吗？',
    yes: '是',
    no: '否'
  },
  menu: {
    dashboard: '仪表板',
    system: '系统管理',
    'system.user': '用户管理',
    'system.role': '角色管理',
    'system.menu': '菜单管理',
    baseData: '基础数据',
    'baseData.factory': '工厂管理',
    'baseData.workshop': '车间管理',
    'baseData.line': '生产线管理',
    'baseData.product': '产品管理',
    'baseData.route': '工艺路线管理',
    'baseData.operation': '工序管理',
    'baseData.recipe': '配方管理',
    workOrder: '工单管理',
    'workOrder.list': '工单列表',
    lot: '批次管理',
    'lot.list': '批次列表',
    wip: 'WIP跟踪',
    'wip.move': '流转记录',
    equipment: '设备管理',
    'equipment.list': '设备台账',
    quality: '质量管理',
    'quality.inspection': '检验记录'
  },
  button: {
    query: '查询',
    add: '新增',
    edit: '编辑',
    delete: '删除'
  },
  auth: {
    login: '登录',
    logout: '退出',
    username: '用户名',
    password: '密码',
    rememberMe: '记住我',
    forgotPassword: '忘记密码',
    welcomeBack: '欢迎回来',
    pleaseLogin: '请登录您的账号'
  },
  user: {
    username: '用户名',
    realName: '真实姓名',
    email: '邮箱',
    phone: '手机号',
    status: '状态',
    roles: '角色',
    createTime: '创建时间',
    updateTime: '更新时间'
  },
  baseData: {
    factory: {
      title: '工厂管理',
      code: '工厂编码',
      name: '工厂名称',
      address: '地址',
      contact: '联系人',
      phone: '联系电话'
    },
    workshop: {
      title: '车间管理',
      code: '车间编码',
      name: '车间名称',
      type: '车间类型',
      factory: '所属工厂',
      description: '描述'
    },
    line: {
      title: '生产线管理',
      code: '生产线编码',
      name: '生产线名称',
      workshop: '所属车间',
      capacity: '产能',
      description: '描述'
    },
    product: {
      title: '产品管理',
      code: '产品编码',
      name: '产品名称',
      type: '产品类型',
      version: '版本',
      description: '描述'
    },
    route: {
      title: '工艺路线管理',
      code: '路线编码',
      name: '路线名称',
      product: '所属产品',
      version: '版本',
      isDefault: '默认路线',
      description: '描述'
    },
    operation: {
      title: '工序管理',
      code: '工序编码',
      name: '工序名称',
      route: '所属路线',
      type: '工序类型',
      sequence: '顺序',
      standardTime: '标准工时（分钟）',
      description: '描述'
    },
    recipe: {
      title: '配方管理',
      code: '配方编码',
      name: '配方名称',
      operation: '所属工序',
      version: '版本',
      parameters: '配方参数',
      isDefault: '默认配方',
      description: '描述'
    }
  }
}
