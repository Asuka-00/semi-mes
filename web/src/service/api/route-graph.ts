import { request } from '../request';

export interface RouteVersion {
  id: number;
  routeID: number;
  versionNo: number;
  state: 'draft' | 'released' | 'obsolete' | string;
  note: string;
}

export interface GraphNode {
  nodeKey: string;
  nodeType: string;
  name: string;
  operationID: number;
  recipeID: number;
  equipmentGroup: string;
  posX: number;
  posY: number;
}

export interface GraphPredicate {
  field: string;
  op: string;
  value: unknown;
}

export interface GraphCondition {
  all?: GraphPredicate[];
  any?: GraphPredicate[];
}

export interface GraphEdge {
  edgeKey: string;
  fromKey: string;
  toKey: string;
  edgeKind: string;
  isDefault: boolean;
  priority: number;
  condition?: GraphCondition | null;
  maxRework: number;
  onExceed: string;
  label: string;
}

export interface GraphIssue {
  code: string;
  message: string;
  ref?: string;
}

export interface ResolveResult {
  action: string;
  nextNodeKey?: string;
  nextNodeType?: string;
  nextNodeName?: string;
  edgeKey?: string;
  reason: string;
}

export function fetchRouteVersions(routeId: number) {
  return request<{ versions: RouteVersion[] }>({ url: `/baseProcessRoute/${routeId}/versions` });
}

export function createRouteVersion(routeId: number, body: { note?: string; copyFrom?: number }) {
  return request<{ version: RouteVersion }>({
    url: `/baseProcessRoute/${routeId}/versions`,
    method: 'post',
    data: body
  });
}

export function fetchRouteGraph(routeId: number, versionId: number) {
  return request<{ version: RouteVersion; nodes: GraphNode[]; edges: GraphEdge[] }>({
    url: `/baseProcessRoute/${routeId}/versions/${versionId}`
  });
}

export function saveRouteGraph(routeId: number, versionId: number, body: { nodes: GraphNode[]; edges: GraphEdge[] }) {
  return request<{ saved: boolean }>({
    url: `/baseProcessRoute/${routeId}/versions/${versionId}/graph`,
    method: 'put',
    data: body
  });
}

export function validateRouteGraph(routeId: number, versionId: number) {
  return request<{ valid: boolean; issues: GraphIssue[] }>({
    url: `/baseProcessRoute/${routeId}/versions/${versionId}/validate`,
    method: 'post',
    data: {}
  });
}

export function releaseRouteGraph(routeId: number, versionId: number) {
  return request<{ released: boolean; version: RouteVersion; issues: GraphIssue[] }>({
    url: `/baseProcessRoute/${routeId}/versions/${versionId}/release`,
    method: 'post',
    data: {}
  });
}

export function resolveRouteStep(
  routeId: number,
  versionId: number,
  body: {
    currentNodeKey: string;
    context: {
      inspection?: { result?: string; grade?: string };
      defect?: { code?: string };
      lot?: { productCode?: string; priority?: number; type?: string };
      reworkCounts?: Record<string, number>;
    };
  }
) {
  return request<ResolveResult>({
    url: `/baseProcessRoute/${routeId}/versions/${versionId}/resolve`,
    method: 'post',
    data: body
  });
}
