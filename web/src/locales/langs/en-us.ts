const local: App.I18n.Schema = {
  system: {
    title: 'Semiconductor MES',
    updateTitle: 'System Version Update Notification',
    updateContent: 'A new version of the system has been detected. Do you want to refresh the page immediately?',
    updateConfirm: 'Refresh immediately',
    updateCancel: 'Later'
  },
  common: {
    action: 'Action',
    add: 'Add',
    addSuccess: 'Add Success',
    backToHome: 'Back to home',
    batchDelete: 'Batch Delete',
    cancel: 'Cancel',
    close: 'Close',
    check: 'Check',
    selectAll: 'Select All',
    expandColumn: 'Expand Column',
    columnSetting: 'Column Setting',
    config: 'Config',
    confirm: 'Confirm',
    delete: 'Delete',
    deleteSuccess: 'Delete Success',
    confirmDelete: 'Are you sure you want to delete?',
    edit: 'Edit',
    warning: 'Warning',
    error: 'Error',
    index: 'Index',
    keywordSearch: 'Please enter keyword',
    logout: 'Logout',
    logoutConfirm: 'Are you sure you want to log out?',
    lookForward: 'Coming soon',
    modify: 'Modify',
    modifySuccess: 'Modify Success',
    noData: 'No Data',
    operate: 'Operate',
    pleaseCheckValue: 'Please check whether the value is valid',
    refresh: 'Refresh',
    reset: 'Reset',
    search: 'Search',
    switch: 'Switch',
    tip: 'Tip',
    trigger: 'Trigger',
    update: 'Update',
    updateSuccess: 'Update Success',
    userCenter: 'User Center',
    yesOrNo: {
      yes: 'Yes',
      no: 'No'
    }
  },
  request: {
    logout: 'Logout user after request failed',
    logoutMsg: 'User status is invalid, please log in again',
    logoutWithModal: 'Pop up modal after request failed and then log out user',
    logoutWithModalMsg: 'User status is invalid, please log in again',
    refreshToken: 'The requested token has expired, refresh the token',
    tokenExpired: 'The requested token has expired'
  },
  theme: {
    themeDrawerTitle: 'Theme Configuration',
    tabs: {
      appearance: 'Appearance',
      layout: 'Layout',
      general: 'General',
      preset: 'Preset'
    },
    appearance: {
      themeSchema: {
        title: 'Theme Schema',
        light: 'Light',
        dark: 'Dark',
        auto: 'Follow System'
      },
      grayscale: 'Grayscale',
      colourWeakness: 'Colour Weakness',
      themeColor: {
        title: 'Theme Color',
        primary: 'Primary',
        info: 'Info',
        success: 'Success',
        warning: 'Warning',
        error: 'Error',
        followPrimary: 'Follow Primary'
      },
      themeRadius: {
        title: 'Theme Radius'
      },
      recommendColor: 'Apply Recommended Color Algorithm',
      recommendColorDesc: 'The recommended color algorithm refers to',
      preset: {
        title: 'Theme Presets',
        apply: 'Apply',
        applySuccess: 'Preset applied successfully',
        default: {
          name: 'Default Preset',
          desc: 'Default theme preset with balanced settings'
        },
        dark: {
          name: 'Dark Preset',
          desc: 'Dark theme preset for night time usage'
        },
        compact: {
          name: 'Compact Preset',
          desc: 'Compact layout preset for small screens'
        },
        azir: {
          name: "Azir's Preset",
          desc: 'It is a cold and elegant preset that Azir likes'
        }
      }
    },
    layout: {
      layoutMode: {
        title: 'Layout Mode',
        vertical: 'Vertical Mode',
        horizontal: 'Horizontal Mode',
        'vertical-mix': 'Vertical Mix Mode',
        'vertical-hybrid-header-first': 'Left Hybrid Header-First',
        'top-hybrid-sidebar-first': 'Top-Hybrid Sidebar-First',
        'top-hybrid-header-first': 'Top-Hybrid Header-First',
        vertical_detail: 'Vertical menu layout, with the menu on the left and content on the right.',
        'vertical-mix_detail':
          'Vertical mix-menu layout, with the primary menu on the dark left side and the secondary menu on the lighter left side.',
        'vertical-hybrid-header-first_detail':
          'Left hybrid layout, with the primary menu at the top, the secondary menu on the dark left side, and the tertiary menu on the lighter left side.',
        horizontal_detail: 'Horizontal menu layout, with the menu at the top and content below.',
        'top-hybrid-sidebar-first_detail':
          'Top hybrid layout, with the primary menu on the left and the secondary menu at the top.',
        'top-hybrid-header-first_detail':
          'Top hybrid layout, with the primary menu at the top and the secondary menu on the left.'
      },
      tab: {
        title: 'Tab Settings',
        visible: 'Tab Visible',
        cache: 'Tag Bar Info Cache',
        cacheTip: 'Keep the tab bar information after leaving the page',
        height: 'Tab Height',
        mode: {
          title: 'Tab Mode',
          slider: 'Slider',
          chrome: 'Chrome',
          button: 'Button'
        },
        closeByMiddleClick: 'Close Tab by Middle Click',
        closeByMiddleClickTip: 'Enable closing tabs by clicking with the middle mouse button'
      },
      header: {
        title: 'Header Settings',
        height: 'Header Height',
        breadcrumb: {
          visible: 'Breadcrumb Visible',
          showIcon: 'Breadcrumb Icon Visible'
        }
      },
      sider: {
        title: 'Sider Settings',
        inverted: 'Dark Sider',
        width: 'Sider Width',
        collapsedWidth: 'Sider Collapsed Width',
        mixWidth: 'Mix Sider Width',
        mixCollapsedWidth: 'Mix Sider Collapse Width',
        mixChildMenuWidth: 'Mix Child Menu Width',
        autoSelectFirstMenu: 'Auto Select First Submenu',
        autoSelectFirstMenuTip:
          'When a first-level menu is clicked, the first submenu is automatically selected and navigated to the deepest level'
      },
      footer: {
        title: 'Footer Settings',
        visible: 'Footer Visible',
        fixed: 'Fixed Footer',
        height: 'Footer Height',
        right: 'Right Footer'
      },
      content: {
        title: 'Content Area Settings',
        scrollMode: {
          title: 'Scroll Mode',
          tip: 'The theme scroll only scrolls the main part, the outer scroll can carry the header and footer together',
          wrapper: 'Wrapper',
          content: 'Content'
        },
        page: {
          animate: 'Page Animate',
          mode: {
            title: 'Page Animate Mode',
            fade: 'Fade',
            'fade-slide': 'Slide',
            'fade-bottom': 'Fade Zoom',
            'fade-scale': 'Fade Scale',
            'zoom-fade': 'Zoom Fade',
            'zoom-out': 'Zoom Out',
            none: 'None'
          }
        },
        fixedHeaderAndTab: 'Fixed Header And Tab'
      }
    },
    general: {
      title: 'General Settings',
      watermark: {
        title: 'Watermark Settings',
        visible: 'Watermark Full Screen Visible',
        text: 'Custom Watermark Text',
        enableUserName: 'Enable User Name Watermark',
        enableTime: 'Show Current Time',
        timeFormat: 'Time Format'
      },
      multilingual: {
        title: 'Multilingual Settings',
        visible: 'Display multilingual button'
      },
      globalSearch: {
        title: 'Global Search Settings',
        visible: 'Display GlobalSearch button'
      }
    },
    configOperation: {
      copyConfig: 'Copy Config',
      copySuccessMsg: 'Copy Success, Please replace the variable "themeSettings" in "src/theme/settings.ts"',
      resetConfig: 'Reset Config',
      resetSuccessMsg: 'Reset Success'
    }
  },
  route: {
    login: 'Login',
    403: 'No Permission',
    404: 'Page Not Found',
    500: 'Server Error',
    'iframe-page': 'Iframe',
    home: 'Home',
    system: 'System',
    system_user: 'Users',
    system_role: 'Roles',
    system_menu: 'Menus',
    'base-data': 'Base Data',
    'base-data_factory': 'Factory',
    'base-data_workshop': 'Workshop',
    'base-data_line': 'Production Line',
    'base-data_product': 'Product',
    'base-data_route': 'Process Route',
    'base-data_operation': 'Operation',
    'base-data_recipe': 'Recipe',
    'work-order': 'Work Order',
    'work-order_list': 'Work Orders',
    lot: 'Lot',
    lot_list: 'Lots',
    wip: 'WIP',
    wip_move: 'Move History',
    equipment: 'Equipment',
    equipment_list: 'Equipment Ledger',
    quality: 'Quality',
    quality_inspection: 'Inspections'
  },
  page: {
    login: {
      common: {
        loginOrRegister: 'Login / Register',
        userNamePlaceholder: 'Please enter user name',
        phonePlaceholder: 'Please enter phone number',
        codePlaceholder: 'Please enter verification code',
        passwordPlaceholder: 'Please enter password',
        confirmPasswordPlaceholder: 'Please enter password again',
        codeLogin: 'Verification code login',
        confirm: 'Confirm',
        back: 'Back',
        validateSuccess: 'Verification passed',
        loginSuccess: 'Login successfully',
        welcomeBack: 'Welcome back, {userName} !'
      },
      pwdLogin: {
        title: 'Password Login',
        rememberMe: 'Remember me',
        forgetPassword: 'Forget password?',
        register: 'Register',
        otherAccountLogin: 'Other Account Login',
        otherLoginMode: 'Other Login Mode',
        superAdmin: 'Super Admin',
        admin: 'Admin',
        user: 'User'
      },
      codeLogin: {
        title: 'Verification Code Login',
        getCode: 'Get verification code',
        reGetCode: 'Reacquire after {time}s',
        sendCodeSuccess: 'Verification code sent successfully',
        imageCodePlaceholder: 'Please enter image verification code'
      },
      register: {
        title: 'Register',
        agreement: 'I have read and agree to',
        protocol: '《User Agreement》',
        policy: '《Privacy Policy》'
      },
      resetPwd: {
        title: 'Reset Password'
      },
      bindWeChat: {
        title: 'Bind WeChat'
      }
    },
    home: {
      branchDesc:
        'For the convenience of everyone in developing and updating the merge, we have streamlined the code of the main branch, only retaining the homepage menu, and the rest of the content has been moved to the example branch for maintenance. The preview address displays the content of the example branch.',
      greeting: 'Good morning, {userName}, today is another day full of vitality!',
      weatherDesc: 'Today is cloudy to clear, 20℃ - 25℃!',
      projectCount: 'Project Count',
      todo: 'Todo',
      message: 'Message',
      downloadCount: 'Download Count',
      registerCount: 'Register Count',
      schedule: 'Work and rest Schedule',
      study: 'Study',
      work: 'Work',
      rest: 'Rest',
      entertainment: 'Entertainment',
      visitCount: 'Visit Count',
      turnover: 'Turnover',
      dealCount: 'Deal Count',
      projectNews: {
        title: 'Project News',
        moreNews: 'More News',
        desc1: 'Soybean created the open source project soybean-admin on May 28, 2021!',
        desc2: 'Yanbowe submitted a bug to soybean-admin, the multi-tab bar will not adapt.',
        desc3: 'Soybean is ready to do sufficient preparation for the release of soybean-admin!',
        desc4: 'Soybean is busy writing project documentation for soybean-admin!',
        desc5: 'Soybean just wrote some of the workbench pages casually, and it was enough to see!'
      },
      creativity: 'Creativity'
    },
    mes: {
      enabled: 'Enabled',
      disabled: 'Disabled',
      yes: 'Yes',
      no: 'No',
      selectPlaceholder: 'Please select',
      operations: 'Operations',
      moveUp: 'Move up',
      moveDown: 'Move down',
      assignRoles: 'Assign roles',
      assignMenus: 'Assign menus',
      passwordOptional: 'Leave blank to keep the current password',
      root: 'Root',
      directory: 'Directory',
      menuPage: 'Page',
      button: 'Button',
      selectRecord: 'Select a process route to maintain its operation sequence',
      field: {
        factoryCode: 'Factory code',
        factoryName: 'Factory name',
        address: 'Address',
        contact: 'Contact',
        phone: 'Phone',
        status: 'Status',
        workshopCode: 'Workshop code',
        workshopName: 'Workshop name',
        workshopType: 'Workshop type',
        factory: 'Factory',
        description: 'Description',
        lineCode: 'Line code',
        lineName: 'Line name',
        capacity: 'Capacity',
        workshop: 'Workshop',
        productCode: 'Product code',
        productName: 'Product name',
        productType: 'Product type',
        version: 'Version',
        routeCode: 'Route code',
        routeName: 'Route name',
        product: 'Product',
        isDefault: 'Default',
        operationCode: 'Operation code',
        operationName: 'Operation name',
        operationType: 'Operation type',
        sequence: 'Sequence',
        standardTime: 'Standard time',
        route: 'Process route',
        recipeCode: 'Recipe code',
        recipeName: 'Recipe name',
        parameters: 'Parameters',
        operation: 'Operation',
        username: 'Username',
        password: 'Password',
        realName: 'Name',
        email: 'Email',
        roleCode: 'Role code',
        roleName: 'Role name',
        menuName: 'Menu name',
        menuType: 'Menu type',
        permissionCode: 'Permission code',
        menuRouteName: 'Route name',
        menuRoutePath: 'Route path',
        menuComponentPath: 'Component',
        icon: 'Icon',
        sortOrder: 'Sort',
        parent: 'Parent menu'
      },
      routeGraph: {
        createRoute: 'New route',
        editRoute: 'Edit route',
        deleteRoute: 'Delete route',
        selectRoute: 'Select a process route to edit its flowchart',
        versions: 'Version',
        stateDraft: 'Draft',
        stateReleased: 'Released',
        stateObsolete: 'Obsolete',
        newDraft: 'Blank draft',
        copyDraft: 'Copy as new draft',
        save: 'Save draft',
        validate: 'Validate',
        release: 'Release',
        autoLayout: 'Auto layout',
        readonlyHint: 'Released and obsolete versions are read-only. Copy a draft to edit.',
        palette: 'Drag to add',
        addStart: 'Start',
        addEnd: 'End',
        addOperation: 'Operation',
        addDecision: 'Decision',
        dragHint: 'Drop onto the canvas, or select a node and connect it from the side panel',
        nodePanel: 'Node',
        edgePanel: 'Edge',
        emptySelection: 'Select a node or edge to edit it here',
        nodeName: 'Name',
        equipmentGroup: 'Equipment group',
        edgeKind: 'Edge kind',
        kindNormal: 'Normal',
        kindRework: 'Rework',
        defaultEdge: 'Default path',
        priority: 'Priority',
        maxRework: 'Max rework count',
        onExceedHold: 'Hold when exceeded',
        conditionMode: 'Condition match',
        matchAll: 'Match all',
        matchAny: 'Match any',
        addPredicate: 'Add condition',
        field: 'Field',
        op: 'Operator',
        value: 'Value',
        deleteNode: 'Delete node',
        deleteEdge: 'Delete edge',
        connectTo: 'Connect to',
        validOk: 'Validation passed',
        releasedOk: 'Released',
        savedOk: 'Draft saved',
        issueTitle: 'Release checks failed',
        rework: 'Rework',
        nodeStart: 'Start',
        nodeEnd: 'End',
        nodeOperation: 'Operation',
        nodeDecision: 'Decision',
        fieldInspectionResult: 'Inspection result',
        fieldInspectionGrade: 'Inspection grade',
        fieldDefectCode: 'Defect code',
        fieldProductCode: 'Product code',
        fieldPriority: 'Priority',
        fieldLotType: 'Lot type',
        fieldReworkCount: 'Rework count on this edge',
        opEq: 'Equals',
        opNe: 'Not equals',
        opGt: 'Greater than',
        opGte: 'Greater or equal',
        opLt: 'Less than',
        opLte: 'Less or equal',
        opIn: 'In',
        opContains: 'Contains',
        inHint: 'Separate multiple values with commas',
        noVersion: 'No version yet',
        releaseBlocked: 'Fix validation issues before release',
        simulate: 'Try next step',
        currentNode: 'Current node',
        inspectionResult: 'Inspection result',
        lotType: 'Lot type',
        reworkCount: 'Rework count taken',
        runResolve: 'Resolve',
        pass: 'Pass',
        fail: 'Fail',
        production: 'Production',
        engineering: 'Engineering',
        resultMove: 'Move',
        resultHold: 'Hold',
        resultEnd: 'End',
        reasonMatched: 'Condition matched',
        reasonDefault: 'Default path',
        reasonRework: 'Rework limit exceeded',
        reasonEnd: 'Already at end',
        recipe: 'Recipe',
        none: 'None'
      }
    }
  },
  form: {
    required: 'Cannot be empty',
    userName: {
      required: 'Please enter user name',
      invalid: 'User name format is incorrect'
    },
    phone: {
      required: 'Please enter phone number',
      invalid: 'Phone number format is incorrect'
    },
    pwd: {
      required: 'Please enter password',
      invalid: '6-18 characters, including letters, numbers, and underscores'
    },
    confirmPwd: {
      required: 'Please enter password again',
      invalid: 'The two passwords are inconsistent'
    },
    code: {
      required: 'Please enter verification code',
      invalid: 'Verification code format is incorrect'
    },
    email: {
      required: 'Please enter email',
      invalid: 'Email format is incorrect'
    }
  },
  dropdown: {
    closeCurrent: 'Close Current',
    closeOther: 'Close Other',
    closeLeft: 'Close Left',
    closeRight: 'Close Right',
    closeAll: 'Close All',
    pin: 'Pin Tab',
    unpin: 'Unpin Tab'
  },
  icon: {
    themeConfig: 'Theme Configuration',
    themeSchema: 'Theme Schema',
    lang: 'Switch Language',
    fullscreen: 'Fullscreen',
    fullscreenExit: 'Exit Fullscreen',
    reload: 'Reload Page',
    collapse: 'Collapse Menu',
    expand: 'Expand Menu',
    pin: 'Pin',
    unpin: 'Unpin'
  },
  datatable: {
    itemCount: 'Total {total} items',
    fixed: {
      left: 'Left Fixed',
      right: 'Right Fixed',
      unFixed: 'Unfixed'
    }
  }
};

export default local;
