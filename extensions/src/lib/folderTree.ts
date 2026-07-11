export interface FolderNode {
  id: string;
  name: string;
  parent_id: string | null;
  children: FolderNode[];
}

export function buildFolderTree(
  folders: Array<{ id: string; name: string; parent_id: string | null }>,
): FolderNode[] {
  const nodeMap = new Map<string, FolderNode>();
  const visited = new Set<string>();
  const roots: FolderNode[] = [];

  for (const f of folders) {
    nodeMap.set(f.id, { ...f, children: [] });
  }

  for (const f of folders) {
    const node = nodeMap.get(f.id)!;
    if (f.parent_id && nodeMap.has(f.parent_id) && !visited.has(f.id)) {
      visited.add(f.id);
      nodeMap.get(f.parent_id)!.children.push(node);
    } else if (!f.parent_id) {
      roots.push(node);
    }
  }

  return roots;
}

export function filterFolderTree(nodes: FolderNode[], term: string): FolderNode[] {
  const lowerTerm = term.toLowerCase();
  const result: FolderNode[] = [];

  for (const node of nodes) {
    const filteredChildren = filterFolderTree(node.children, term);
    const selfMatches = node.name.toLowerCase().includes(lowerTerm);
    if (selfMatches || filteredChildren.length > 0) {
      result.push({ ...node, children: filteredChildren });
    }
  }

  return result;
}
