import { useEffect, useState } from "react";
import { getFolders } from "../lib/api";
import { getDefaultFolder, setDefaultFolder, clearDefaultFolder } from "../lib/storage";
import { buildFolderTree, filterFolderTree, type FolderNode } from "../lib/folderTree";
import type { Colors } from "./App";

function FolderRow({
  node,
  depth,
  defaultFolderId,
  onSelect,
  c,
}: {
  node: FolderNode;
  depth: number;
  defaultFolderId: string | null;
  onSelect: (id: string, name: string) => void;
  c: Colors;
}) {
  const isSelected = node.id === defaultFolderId;
  return (
    <>
      <div
        onClick={() => onSelect(node.id, node.name)}
        style={{
          height: 36,
          display: "flex",
          alignItems: "center",
          paddingLeft: 14 + depth * 12,
          paddingRight: 14,
          gap: 8,
          cursor: "pointer",
        }}
      >
        <span
          style={{
            fontSize: 13,
            color: c.text,
            flex: 1,
            overflow: "hidden",
            textOverflow: "ellipsis",
            whiteSpace: "nowrap",
            fontWeight: isSelected ? 600 : 400,
          }}
        >
          {node.name}
        </span>
        {isSelected && (
          <span style={{ fontSize: 12, color: c.accent }}>✓</span>
        )}
      </div>
      {node.children.map((child) => (
        <FolderRow
          key={child.id}
          node={child}
          depth={depth + 1}
          defaultFolderId={defaultFolderId}
          onSelect={onSelect}
          c={c}
        />
      ))}
    </>
  );
}

export default function FolderPicker({
  onBack,
  c,
}: {
  onBack: () => void;
  c: Colors;
}) {
  const [tree, setTree] = useState<FolderNode[]>([]);
  const [loading, setLoading] = useState(true);
  const [fetchError, setFetchError] = useState(false);
  const [filter, setFilter] = useState("");
  const [defaultFolderId, setDefaultFolderIdState] = useState<string | null>(null);

  useEffect(() => {
    getDefaultFolder().then((folder) => {
      setDefaultFolderIdState(folder?.id ?? null);
    });
    getFolders()
      .then((folders) => {
        setTree(buildFolderTree(folders));
      })
      .catch(() => {
        setFetchError(true);
      })
      .finally(() => {
        setLoading(false);
      });
  }, []);

  async function handleSelect(id: string, name: string) {
    await setDefaultFolder({ id, name });
    onBack();
  }

  async function handleSelectNone() {
    await clearDefaultFolder();
    onBack();
  }

  const visibleTree = filter.trim() ? filterFolderTree(tree, filter) : tree;
  const noneIsSelected = defaultFolderId === null;

  return (
    <div
      style={{
        width: 320,
        background: c.bg,
        border: `1px solid ${c.border}`,
        fontFamily:
          '-apple-system, BlinkMacSystemFont, "Segoe UI", "Helvetica Neue", sans-serif',
        overflow: "hidden",
      }}
    >
      {/* Header */}
      <div
        style={{
          height: 44,
          display: "flex",
          alignItems: "center",
          padding: "0 14px",
          borderBottom: `1px solid ${c.border}`,
          gap: 9,
        }}
      >
        <button
          onClick={onBack}
          title="Back"
          style={{
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            width: 22,
            height: 22,
            color: c.textSec,
            background: "transparent",
            border: "none",
            cursor: "pointer",
            padding: 0,
            fontSize: 15,
          }}
        >
          ←
        </button>
        <span
          style={{
            fontSize: 14,
            fontWeight: 600,
            color: c.text,
            letterSpacing: "-0.02em",
          }}
        >
          Default folder
        </span>
      </div>

      {/* Filter input */}
      <div style={{ padding: "10px 14px 6px" }}>
        <input
          type="text"
          placeholder="Search folders…"
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
          disabled={loading}
          style={{
            width: "100%",
            boxSizing: "border-box",
            fontSize: 13,
            color: c.text,
            background: c.bg,
            border: `1px solid ${c.border}`,
            borderRadius: 6,
            padding: "6px 10px",
            outline: "none",
            opacity: loading ? 0.5 : 1,
          }}
        />
      </div>

      {/* Folder rows */}
      <div style={{ overflowY: "auto", maxHeight: 400 }}>
        {/* None (Unsorted) row — always shown */}
        <div
          onClick={handleSelectNone}
          style={{
            height: 36,
            display: "flex",
            alignItems: "center",
            paddingLeft: 14,
            paddingRight: 14,
            gap: 8,
            cursor: "pointer",
          }}
        >
          <span
            style={{
              fontSize: 13,
              color: c.textSec,
              flex: 1,
              fontWeight: noneIsSelected ? 600 : 400,
            }}
          >
            None (Unsorted)
          </span>
          {noneIsSelected && (
            <span style={{ fontSize: 12, color: c.accent }}>✓</span>
          )}
        </div>

        {/* Tree rows */}
        {!fetchError &&
          visibleTree.map((node) => (
            <FolderRow
              key={node.id}
              node={node}
              depth={0}
              defaultFolderId={defaultFolderId}
              onSelect={handleSelect}
              c={c}
            />
          ))}
      </div>
    </div>
  );
}
