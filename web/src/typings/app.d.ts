/** The global namespace for the app */
declare namespace App {
  /** Theme namespace */
  namespace Theme {
    type ColorPaletteNumber = import('@sa/color').ColorPaletteNumber;

    /** NaiveUI theme overrides that can be specified in preset */
    type NaiveUIThemeOverride = import('naive-ui').GlobalThemeOverrides;

    /** Theme setting */
    interface ThemeSetting {
      /** Theme scheme */
      themeScheme: UnionKey.ThemeScheme;
      /** grayscale mode */
      grayscale: boolean;
      /** colour weakness mode */
      colourWeakness: boolean;
      /** Whether to recommend color */
      recommendColor: boolean;
      /** Theme color */
      themeColor: string;
      /** Theme radius */
      themeRadius: number;
      /** Other color */
      otherColor: OtherColor;
      /** Whether info color is followed by the primary color */
      isInfoFollowPrimary: boolean;
      /** Layout */
      layout: {
        /** Layout mode */
        mode: UnionKey.ThemeLayoutMode;
        /** Scroll mode */
        scrollMode: UnionKey.ThemeScrollMode;
      };
      /** Page */
      page: {
        /** Whether to show the page transition */
        animate: boolean;
        /** Page animate mode */
        animateMode: UnionKey.ThemePageAnimateMode;
      };
      /** Header */
      header: {
        /** Header height */
        height: number;
        /** Header breadcrumb */
        breadcrumb: {
          /** Whether to show the breadcrumb */
          visible: boolean;
          /** Whether to show the breadcrumb icon */
          showIcon: boolean;
        };
        /** Multilingual */
        multilingual: {
          /** Whether to show the multilingual */
          visible: boolean;
        };
        globalSearch: {
          /** Whether to show the GlobalSearch */
          visible: boolean;
        };
      };
      /** Tab */
      tab: {
        /** Whether to show the tab */
        visible: boolean;
        /**
         * Whether to cache the tab
         *
         * If cache, the tabs will get from the local storage when the page is refreshed
         */
        cache: boolean;
        /** Tab height */
        height: number;
        /** Tab mode */
        mode: UnionKey.ThemeTabMode;
        /** Whether to close tab by middle click */
        closeTabByMiddleClick: boolean;
      };
      /** Fixed header and tab */
      fixedHeaderAndTab: boolean;
      /** Sider */
      sider: {
        /** Inverted sider */
        inverted: boolean;
        /** Sider width */
        width: number;
        /** Collapsed sider width */
        collapsedWidth: number;
        /** Sider width when the layout is 'vertical-mix', 'top-hybrid-sidebar-first', or 'top-hybrid-header-first' */
        mixWidth: number;
        /**
         * Collapsed sider width when the layout is 'vertical-mix', 'top-hybrid-sidebar-first', or
         * 'top-hybrid-header-first'
         */
        mixCollapsedWidth: number;
        /** Child menu width when the layout is 'vertical-mix', 'top-hybrid-sidebar-first', or 'top-hybrid-header-first' */
        mixChildMenuWidth: number;
        /** Whether to auto select the first submenu */
        autoSelectFirstMenu: boolean;
      };
      /** Footer */
      footer: {
        /** Whether to show the footer */
        visible: boolean;
        /** Whether fixed the footer */
        fixed: boolean;
        /** Footer height */
        height: number;
        /**
         * Whether float the footer to the right when the layout is 'top-hybrid-sidebar-first' or
         * 'top-hybrid-header-first'
         */
        right: boolean;
      };
      /** Watermark */
      watermark: {
        /** Whether to show the watermark */
        visible: boolean;
        /** Watermark text */
        text: string;
        /** Whether to use user name as watermark text */
        enableUserName: boolean;
        /** Whether to use current time as watermark text */
        enableTime: boolean;
        /** Time format for watermark text */
        timeFormat: string;
      };
      /** define some theme settings tokens, will transform to css variables */
      tokens: {
        light: ThemeSettingToken;
        dark?: {
          [K in keyof ThemeSettingToken]?: Partial<ThemeSettingToken[K]>;
        };
      };
    }

    interface OtherColor {
      info: string;
      success: string;
      warning: string;
      error: string;
    }

    interface ThemeColor extends OtherColor {
      primary: string;
    }

    type ThemeColorKey = keyof ThemeColor;

    type ThemePaletteColor = {
      [key in ThemeColorKey | `${ThemeColorKey}-${ColorPaletteNumber}`]: string;
    };

    type BaseToken = Record<string, Record<string, string>>;

    interface ThemeSettingTokenColor {
      /** the progress bar color, if not set, will use the primary color */
      nprogress?: string;
      container: string;
      layout: string;
      inverted: string;
      'base-text': string;
    }

    interface ThemeSettingTokenBoxShadow {
      header: string;
      sider: string;
      tab: string;
    }

    interface ThemeSettingToken {
      colors: ThemeSettingTokenColor;
      boxShadow: ThemeSettingTokenBoxShadow;
    }

    type ThemeTokenColor = ThemePaletteColor & ThemeSettingTokenColor;

    /** Theme token CSS variables */
    type ThemeTokenCSSVars = {
      colors: ThemeTokenColor & { [key: string]: string };
      boxShadow: ThemeSettingTokenBoxShadow & { [key: string]: string };
    };
  }

  /** Global namespace */
  namespace Global {
    type VNode = import('vue').VNode;
    type RouteLocationNormalizedLoaded = import('vue-router').RouteLocationNormalizedLoaded;
    type RouteKey = import('@elegant-router/types').RouteKey;
    type RouteMap = import('@elegant-router/types').RouteMap;
    type RoutePath = import('@elegant-router/types').RoutePath;
    type LastLevelRouteKey = import('@elegant-router/types').LastLevelRouteKey;

    /** The router push options */
    type RouterPushOptions = {
      query?: Record<string, string>;
      params?: Record<string, string>;
      force?: boolean;
    };

    /** The global header props */
    interface HeaderProps {
      /** Whether to show the logo */
      showLogo?: boolean;
      /** Whether to show the menu toggler */
      showMenuToggler?: boolean;
      /** Whether to show the menu */
      showMenu?: boolean;
    }

    /** The global menu */
    type Menu = {
      /**
       * The menu key
       *
       * Equal to the route key
       */
      key: string;
      /** The menu label */
      label: string;
      /** The menu i18n key */
      i18nKey?: I18n.I18nKey | null;
      /** The route key */
      routeKey: RouteKey;
      /** The route path */
      routePath: RoutePath;
      /** The menu icon */
      icon?: () => VNode;
      /** The menu children */
      children?: Menu[];
    };

    type Breadcrumb = Omit<Menu, 'children'> & {
      options?: Breadcrumb[];
    };

    /** Tab route */
    type TabRoute = Pick<RouteLocationNormalizedLoaded, 'name' | 'path' | 'meta'> &
      Partial<Pick<RouteLocationNormalizedLoaded, 'fullPath' | 'query' | 'matched'>>;

    /** The global tab */
    type Tab = {
      /** The tab id */
      id: string;
      /** The tab label */
      label: string;
      /**
       * The new tab label
       *
       * If set, the tab label will be replaced by this value
       */
      newLabel?: string;
      /**
       * The old tab label
       *
       * when reset the tab label, the tab label will be replaced by this value
       */
      oldLabel?: string;
      /** The tab route key */
      routeKey: LastLevelRouteKey;
      /** The tab route path */
      routePath: RouteMap[LastLevelRouteKey];
      /** The tab route full path */
      fullPath: string;
      /** The tab fixed index */
      fixedIndex?: number | null;
      /**
       * Tab icon
       *
       * Iconify icon
       */
      icon?: string;
      /**
       * Tab local icon
       *
       * Local icon
       */
      localIcon?: string;
      /** I18n key */
      i18nKey?: I18n.I18nKey | null;
    };

    /** Form rule */
    type FormRule = import('naive-ui').FormItemRule;

    /** The global dropdown key */
    type DropdownKey = 'closeCurrent' | 'closeOther' | 'closeLeft' | 'closeRight' | 'closeAll' | 'pin' | 'unpin';
  }

  /**
   * I18n namespace
   *
   * Locales type
   */
  namespace I18n {
    type RouteKey = import('@elegant-router/types').RouteKey;

    type LangType = 'en-US' | 'zh-CN';

    type LangOption = {
      label: string;
      key: LangType;
    };

    type I18nRouteKey = Exclude<RouteKey, 'root' | 'not-found'>;

    type FormMsg = {
      required: string;
      invalid: string;
    };

    type Schema = {
      system: {
        title: string;
        updateTitle: string;
        updateContent: string;
        updateConfirm: string;
        updateCancel: string;
      };
      common: {
        action: string;
        add: string;
        addSuccess: string;
        backToHome: string;
        batchDelete: string;
        cancel: string;
        close: string;
        check: string;
        selectAll: string;
        expandColumn: string;
        columnSetting: string;
        config: string;
        confirm: string;
        delete: string;
        deleteSuccess: string;
        confirmDelete: string;
        edit: string;
        warning: string;
        error: string;
        index: string;
        keywordSearch: string;
        logout: string;
        logoutConfirm: string;
        lookForward: string;
        modify: string;
        modifySuccess: string;
        noData: string;
        operate: string;
        pleaseCheckValue: string;
        refresh: string;
        reset: string;
        search: string;
        switch: string;
        tip: string;
        trigger: string;
        update: string;
        updateSuccess: string;
        userCenter: string;
        yesOrNo: {
          yes: string;
          no: string;
        };
      };
      request: {
        logout: string;
        logoutMsg: string;
        logoutWithModal: string;
        logoutWithModalMsg: string;
        refreshToken: string;
        tokenExpired: string;
      };
      theme: {
        themeDrawerTitle: string;
        tabs: {
          appearance: string;
          layout: string;
          general: string;
          preset: string;
        };
        appearance: {
          themeSchema: { title: string } & Record<UnionKey.ThemeScheme, string>;
          grayscale: string;
          colourWeakness: string;
          themeColor: {
            title: string;
            followPrimary: string;
          } & Record<Theme.ThemeColorKey, string>;
          recommendColor: string;
          recommendColorDesc: string;
          themeRadius: {
            title: string;
          };
          preset: {
            title: string;
            apply: string;
            applySuccess: string;
            [key: string]:
              | {
                  name: string;
                  desc: string;
                }
              | string;
          };
        };
        layout: {
          layoutMode: { title: string } & Record<UnionKey.ThemeLayoutMode, string> & {
              [K in `${UnionKey.ThemeLayoutMode}_detail`]: string;
            };
          tab: {
            title: string;
            visible: string;
            cache: string;
            cacheTip: string;
            height: string;
            mode: { title: string } & Record<UnionKey.ThemeTabMode, string>;
            closeByMiddleClick: string;
            closeByMiddleClickTip: string;
          };
          header: {
            title: string;
            height: string;
            breadcrumb: {
              visible: string;
              showIcon: string;
            };
          };
          sider: {
            title: string;
            inverted: string;
            width: string;
            collapsedWidth: string;
            mixWidth: string;
            mixCollapsedWidth: string;
            mixChildMenuWidth: string;
            autoSelectFirstMenu: string;
            autoSelectFirstMenuTip: string;
          };
          footer: {
            title: string;
            visible: string;
            fixed: string;
            height: string;
            right: string;
          };
          content: {
            title: string;
            scrollMode: { title: string; tip: string } & Record<UnionKey.ThemeScrollMode, string>;
            page: {
              animate: string;
              mode: { title: string } & Record<UnionKey.ThemePageAnimateMode, string>;
            };
            fixedHeaderAndTab: string;
          };
        };
        general: {
          title: string;
          watermark: {
            title: string;
            visible: string;
            text: string;
            enableUserName: string;
            enableTime: string;
            timeFormat: string;
          };
          multilingual: {
            title: string;
            visible: string;
          };
          globalSearch: {
            title: string;
            visible: string;
          };
        };
        configOperation: {
          copyConfig: string;
          copySuccessMsg: string;
          resetConfig: string;
          resetSuccessMsg: string;
        };
      };
      route: Record<I18nRouteKey, string>;
      page: {
        login: {
          common: {
            loginOrRegister: string;
            userNamePlaceholder: string;
            phonePlaceholder: string;
            codePlaceholder: string;
            passwordPlaceholder: string;
            confirmPasswordPlaceholder: string;
            codeLogin: string;
            confirm: string;
            back: string;
            validateSuccess: string;
            loginSuccess: string;
            welcomeBack: string;
          };
          pwdLogin: {
            title: string;
            rememberMe: string;
            forgetPassword: string;
            register: string;
            otherAccountLogin: string;
            otherLoginMode: string;
            superAdmin: string;
            admin: string;
            user: string;
          };
          codeLogin: {
            title: string;
            getCode: string;
            reGetCode: string;
            sendCodeSuccess: string;
            imageCodePlaceholder: string;
          };
          register: {
            title: string;
            agreement: string;
            protocol: string;
            policy: string;
          };
          resetPwd: {
            title: string;
          };
          bindWeChat: {
            title: string;
          };
        };
        home: {
          branchDesc: string;
          greeting: string;
          weatherDesc: string;
          projectCount: string;
          todo: string;
          message: string;
          downloadCount: string;
          registerCount: string;
          schedule: string;
          study: string;
          work: string;
          rest: string;
          entertainment: string;
          visitCount: string;
          turnover: string;
          dealCount: string;
          projectNews: {
            title: string;
            moreNews: string;
            desc1: string;
            desc2: string;
            desc3: string;
            desc4: string;
            desc5: string;
          };
          creativity: string;
          dash: {
            title: string;
            subtitle: string;
            refresh: string;
            updated: string;
            wipLots: string;
            wipQty: string;
            holdLots: string;
            runningLots: string;
            todayMoves: string;
            todayCompleted: string;
            todayScrap: string;
            byStep: string;
            byStatus: string;
            byProduct: string;
            trend: string;
            moves: string;
            scrap: string;
            lots: string;
            tools: string;
            holds: string;
            duration: string;
            durationValue: string;
            equipment: string;
            overduePm: string;
            notices: string;
            unread: string;
            events: string;
            openOverview: string;
            openBoard: string;
            noWip: string;
            noEquipment: string;
            noNotices: string;
            noEvents: string;
          };
        };
        mes: {
          query: {
            collapse: string;
            expand: string;
            export: string;
            columns: string;
            createdRange: string;
            sort: string;
            batchDelete: string;
            batchEnable: string;
            batchDisable: string;
            batchHold: string;
            batchRelease: string;
            selected: string;
            exported: string;
            exportEmpty: string;
            partial: string;
          };
          audit: {
            title: string;
            user: string;
            module: string;
            action: string;
            target: string;
            entityType: string;
            ip: string;
            time: string;
            summary: string;
            diff: string;
            before: string;
            after: string;
            emptyDiff: string;
          };
          trace: {
            title: string;
            lotNo: string;
            load: string;
            forward: string;
            backward: string;
            emptyTree: string;
            history: string;
            moves: string;
            defects: string;
            measurements: string;
            reverse: string;
            equipment: string;
            node: string;
            from: string;
            to: string;
            searchReverse: string;
            print: string;
            export: string;
            operator: string;
            recipe: string;
            params: string;
            related: string;
            reason: string;
            value: string;
            disposition: string;
            wafersLater: string;
            notFound: string;
          };
          enabled: string;
          disabled: string;
          yes: string;
          no: string;
          selectPlaceholder: string;
          operations: string;
          moveUp: string;
          moveDown: string;
          assignRoles: string;
          assignMenus: string;
          passwordOptional: string;
          root: string;
          directory: string;
          menuPage: string;
          button: string;
          selectRecord: string;
          field: {
            factoryCode: string;
            factoryName: string;
            address: string;
            contact: string;
            phone: string;
            status: string;
            workshopCode: string;
            workshopName: string;
            workshopType: string;
            factory: string;
            description: string;
            lineCode: string;
            lineName: string;
            capacity: string;
            workshop: string;
            productCode: string;
            productName: string;
            productType: string;
            version: string;
            routeCode: string;
            routeName: string;
            product: string;
            isDefault: string;
            operationCode: string;
            operationName: string;
            operationType: string;
            sequence: string;
            standardTime: string;
            route: string;
            recipeCode: string;
            recipeName: string;
            parameters: string;
            operation: string;
            username: string;
            password: string;
            realName: string;
            email: string;
            roleCode: string;
            roleName: string;
            menuName: string;
            menuType: string;
            permissionCode: string;
            menuRouteName: string;
            menuRoutePath: string;
            menuComponentPath: string;
            icon: string;
            sortOrder: string;
            parent: string;
          };
          routeGraph: {
            createRoute: string;
            editRoute: string;
            deleteRoute: string;
            selectRoute: string;
            versions: string;
            stateDraft: string;
            stateReleased: string;
            stateObsolete: string;
            newDraft: string;
            copyDraft: string;
            save: string;
            validate: string;
            release: string;
            autoLayout: string;
            readonlyHint: string;
            palette: string;
            addStart: string;
            addEnd: string;
            addOperation: string;
            addDecision: string;
            dragHint: string;
            nodePanel: string;
            edgePanel: string;
            emptySelection: string;
            nodeName: string;
            equipmentGroup: string;
            edgeKind: string;
            kindNormal: string;
            kindRework: string;
            defaultEdge: string;
            priority: string;
            maxRework: string;
            onExceedHold: string;
            conditionMode: string;
            matchAll: string;
            matchAny: string;
            addPredicate: string;
            field: string;
            op: string;
            value: string;
            deleteNode: string;
            deleteEdge: string;
            connectTo: string;
            validOk: string;
            releasedOk: string;
            savedOk: string;
            issueTitle: string;
            rework: string;
            nodeStart: string;
            nodeEnd: string;
            nodeOperation: string;
            nodeDecision: string;
            fieldInspectionResult: string;
            fieldInspectionGrade: string;
            fieldDefectCode: string;
            fieldProductCode: string;
            fieldPriority: string;
            fieldLotType: string;
            fieldReworkCount: string;
            opEq: string;
            opNe: string;
            opGt: string;
            opGte: string;
            opLt: string;
            opLte: string;
            opIn: string;
            opContains: string;
            inHint: string;
            noVersion: string;
            releaseBlocked: string;
            simulate: string;
            currentNode: string;
            inspectionResult: string;
            lotType: string;
            reworkCount: string;
            runResolve: string;
            pass: string;
            fail: string;
            production: string;
            engineering: string;
            resultMove: string;
            resultHold: string;
            resultEnd: string;
            reasonMatched: string;
            reasonDefault: string;
            reasonRework: string;
            reasonEnd: string;
            recipe: string;
            none: string;
          };
          wip: {
            orderNo: string;
            product: string;
            routeVersion: string;
            plannedQty: string;
            releasedQty: string;
            completedQty: string;
            priority: string;
            dueDate: string;
            status: string;
            note: string;
            created: string;
            released: string;
            inProgress: string;
            completed: string;
            closed: string;
            waiting: string;
            hold: string;
            merged: string;
            releaseOrder: string;
            closeOrder: string;
            startLot: string;
            quantity: string;
            lotType: string;
            lotNo: string;
            currentNode: string;
            holdAction: string;
            releaseHold: string;
            reasonCode: string;
            reason: string;
            split: string;
            splitHint: string;
            merge: string;
            advance: string;
            inspectionResult: string;
            defectCode: string;
            detail: string;
            history: string;
            genealogy: string;
            position: string;
            back: string;
            event: string;
            fromNode: string;
            toNode: string;
            priorityLow: string;
            priorityNormal: string;
            priorityHigh: string;
            priorityUrgent: string;
            production: string;
            engineering: string;
            selectVersion: string;
            saved: string;
            noReleased: string;
            linkSplit: string;
            linkMerge: string;
            eventStart: string;
            eventAdvance: string;
            eventHold: string;
            eventRelease: string;
            eventSplit: string;
            eventMerge: string;
            eventComplete: string;
            eventTrackIn: string;
            eventTrackOut: string;
            eventAbort: string;
            eventPass: string;
            eventScrap: string;
            result: string;
            running: string;
            scrapped: string;
            moves: string;
          };
          track: {
            station: string;
            lotNo: string;
            load: string;
            step: string;
            equipment: string;
            trackIn: string;
            trackOut: string;
            abort: string;
            pass: string;
            qtyOut: string;
            qtyScrap: string;
            scrapReason: string;
            reasonBroken: string;
            reasonParticle: string;
            reasonScratch: string;
            reasonOther: string;
            queue: string;
            process: string;
            operator: string;
            holdHint: string;
            runningHint: string;
            byStatus: string;
            byNode: string;
            byProduct: string;
            holdList: string;
            count: string;
            qty: string;
            state: string;
            open: string;
            aborted: string;
            idle: string;
            down: string;
            group: string;
            code: string;
            name: string;
            notAllowed: string;
          };
          eqp: {
            code: string;
            name: string;
            group: string;
            type: string;
            model: string;
            vendor: string;
            location: string;
            capacity: string;
            state: string;
            detail: string;
            changeState: string;
            reason: string;
            serial: string;
            line: string;
            chambers: string;
            openLots: string;
            back: string;
            capability: string;
            anyCapability: string;
            anyRecipe: string;
            capabilityHint: string;
            stateHistory: string;
            from: string;
            to: string;
            started: string;
            duration: string;
            recentLots: string;
            pmHistory: string;
            plan: string;
            taskStatus: string;
            result: string;
            utilization: string;
            overdue: string;
            trigger: string;
            intervalDays: string;
            intervalCount: string;
            lotsSince: string;
            block: string;
            equipmentId: string;
            byTime: string;
            byCount: string;
            byBoth: string;
            nextDue: string;
            checklist: string;
            checklistHint: string;
            taskNo: string;
            due: string;
            startPm: string;
            completePm: string;
            state_standby: string;
            state_productive: string;
            state_engineering: string;
            state_scheduled_down: string;
            state_unscheduled_down: string;
            state_non_scheduled: string;
            task_due: string;
            task_overdue: string;
            task_in_progress: string;
            task_done: string;
            task_cancelled: string;
            pass: string;
            fail: string;
          };
          qc: {
            measure: string;
            plan: string;
            defect: string;
            pareto: string;
            spc: string;
            spcOff: string;
            param: string;
            paramName: string;
            unit: string;
            target: string;
            sample: string;
            required: string;
            enabled: string;
            yes: string;
            no: string;
            operation: string;
            product: string;
            latest: string;
            judgement: string;
            noPlan: string;
            code: string;
            name: string;
            category: string;
            severity: string;
            disposition: string;
            useAsIs: string;
            rework: string;
            scrap: string;
            hold: string;
            addCode: string;
            lotId: string;
            saved: string;
            reaction: string;
            kind: string;
            rule: string;
            value: string;
            useJudgement: string;
          };
        };
      };
      form: {
        required: string;
        userName: FormMsg;
        phone: FormMsg;
        pwd: FormMsg;
        confirmPwd: FormMsg;
        code: FormMsg;
        email: FormMsg;
      };
      dropdown: Record<Global.DropdownKey, string>;
      notice: {
        title: string;
        empty: string;
        markAll: string;
      };
      icon: {
        themeConfig: string;
        themeSchema: string;
        lang: string;
        fullscreen: string;
        fullscreenExit: string;
        reload: string;
        collapse: string;
        expand: string;
        pin: string;
        unpin: string;
      };
      datatable: {
        itemCount: string;
        fixed: {
          left: string;
          right: string;
          unFixed: string;
        };
      };
    };

    type GetI18nKey<T extends Record<string, unknown>, K extends keyof T = keyof T> = K extends string
      ? T[K] extends Record<string, unknown>
        ? `${K}.${GetI18nKey<T[K]>}`
        : K
      : never;

    type I18nKey = GetI18nKey<Schema>;

    type TranslateOptions<Locales extends string> = import('vue-i18n').TranslateOptions<Locales>;

    interface $T {
      (key: I18nKey): string;
      (key: I18nKey, plural: number, options?: TranslateOptions<LangType>): string;
      (key: I18nKey, defaultMsg: string, options?: TranslateOptions<I18nKey>): string;
      (key: I18nKey, list: unknown[], options?: TranslateOptions<I18nKey>): string;
      (key: I18nKey, list: unknown[], plural: number): string;
      (key: I18nKey, list: unknown[], defaultMsg: string): string;
      (key: I18nKey, named: Record<string, unknown>, options?: TranslateOptions<LangType>): string;
      (key: I18nKey, named: Record<string, unknown>, plural: number): string;
      (key: I18nKey, named: Record<string, unknown>, defaultMsg: string): string;
    }
  }

  /** Service namespace */
  namespace Service {
    /** Other baseURL key */
    type OtherBaseURLKey = 'demo';

    interface ServiceConfigItem {
      /** The backend service base url */
      baseURL: string;
      /** The proxy pattern of the backend service base url */
      proxyPattern: string;
    }

    interface OtherServiceConfigItem extends ServiceConfigItem {
      key: OtherBaseURLKey;
    }

    /** The backend service config */
    interface ServiceConfig extends ServiceConfigItem {
      /** Other backend service config */
      other: OtherServiceConfigItem[];
    }

    interface SimpleServiceConfig extends Pick<ServiceConfigItem, 'baseURL'> {
      other: Record<OtherBaseURLKey, string>;
    }

    /** The backend service response data */
    type Response<T = unknown> = {
      /** The backend service response code */
      code: string;
      /** The backend service response message */
      msg: string;
      /** The backend service response data */
      data: T;
    };

    /** The demo backend service response data */
    type DemoResponse<T = unknown> = {
      /** The backend service response code */
      status: string;
      /** The backend service response message */
      message: string;
      /** The backend service response data */
      result: T;
    };
  }
}
