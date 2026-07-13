import type { Folder } from "../../types/folder";

export function findFolder(folders: Folder[], id: string): Folder | null {
  for (const folder of folders) {
    if (folder.id === id) return folder;
    const child = findFolder(folder.children ?? [], id);
    if (child) return child;
  }
  return null;
}

export function findFolderPath(folders: Folder[], id: string, parents: Folder[] = []): Folder[] {
  for (const folder of folders) {
    const nextPath = [...parents, folder];
    if (folder.id === id) return nextPath;
    const childPath = findFolderPath(folder.children ?? [], id, nextPath);
    if (childPath.length > 0) return childPath;
  }
  return [];
}

export function getChildFolders(folders: Folder[], currentFolder: Folder | null): Folder[] {
  if (!currentFolder) {
    return folders;
  }
  return currentFolder.children ?? [];
}

export type FlatFolder = Folder & { depth: number };

export function flattenFolders(folders: Folder[], depth = 0, result: FlatFolder[] = []): FlatFolder[] {
  for (const folder of folders) {
    result.push({ ...folder, depth });
    if (folder.children?.length) {
      flattenFolders(folder.children, depth + 1, result);
    }
  }
  return result;
}

export function collectDescendantIds(folder: Folder): string[] {
  const ids: string[] = [];
  const walk = (node: Folder) => {
    for (const child of node.children ?? []) {
      ids.push(child.id);
      walk(child);
    }
  };
  walk(folder);
  return ids;
}

export function canDropFolderOn(draggedId: string, targetId: string | null, folders: Folder[]): boolean {
  if (targetId === draggedId) return false;
  const dragged = findFolder(folders, draggedId);
  if (!dragged) return false;
  if (targetId && collectDescendantIds(dragged).includes(targetId)) return false;
  return true;
}
