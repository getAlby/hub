import { PlusIcon, XIcon } from "lucide-react";
import React from "react";
import { Button } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { useLabelKeySuggestions } from "src/hooks/useLabelKeySuggestions";

const KEY_MAX_LENGTH = 64;
const VALUE_MAX_LENGTH = 1000;
const KEY_SUGGESTIONS_LIST_ID = "transaction-label-key-suggestions";

type Props = {
  initialLabels: Record<string, string> | undefined;
  saving?: boolean;
  onSave: (labels: Record<string, string>) => void | Promise<void>;
  onCancel: () => void;
};

type Row = { key: string; value: string };

function toRows(labels: Record<string, string> | undefined): Row[] {
  if (!labels || Object.keys(labels).length === 0) {
    return [{ key: "", value: "" }];
  }
  return Object.entries(labels).map(([key, value]) => ({ key, value }));
}

function rowsToLabels(rows: Row[]): Record<string, string> {
  const result: Record<string, string> = {};
  for (const row of rows) {
    const key = row.key.trim();
    const value = row.value.trim();
    if (key && value) {
      result[key] = value;
    }
  }
  return result;
}

function TransactionLabelEditor({
  initialLabels,
  saving,
  onSave,
  onCancel,
}: Props) {
  const [rows, setRows] = React.useState<Row[]>(() => toRows(initialLabels));
  const keySuggestions = useLabelKeySuggestions();

  const updateRow = (index: number, patch: Partial<Row>) => {
    setRows((prev) =>
      prev.map((row, i) => (i === index ? { ...row, ...patch } : row))
    );
  };

  const removeRow = (index: number) => {
    setRows((prev) => prev.filter((_, i) => i !== index));
  };

  const addRow = () => {
    setRows((prev) => [...prev, { key: "", value: "" }]);
  };

  const handleSave = () => {
    onSave(rowsToLabels(rows));
  };

  return (
    <div className="flex flex-col gap-3 mt-2 px-1">
      {keySuggestions.length > 0 && (
        <datalist id={KEY_SUGGESTIONS_LIST_ID}>
          {keySuggestions.map((key) => (
            <option key={key} value={key} />
          ))}
        </datalist>
      )}
      {rows.map((row, index) => (
        <div key={index} className="flex items-center gap-2">
          <Input
            placeholder="key"
            value={row.key}
            maxLength={KEY_MAX_LENGTH}
            list={
              keySuggestions.length > 0 ? KEY_SUGGESTIONS_LIST_ID : undefined
            }
            onChange={(e) => updateRow(index, { key: e.target.value })}
            className="w-1/3"
          />
          <Input
            placeholder="value"
            value={row.value}
            maxLength={VALUE_MAX_LENGTH}
            onChange={(e) => updateRow(index, { value: e.target.value })}
          />
          <Button
            type="button"
            variant="ghost"
            size="icon"
            onClick={() => removeRow(index)}
            aria-label="Remove field"
          >
            <XIcon className="size-4" />
          </Button>
        </div>
      ))}
      <Button
        type="button"
        variant="outline"
        size="sm"
        onClick={addRow}
        className="self-start"
      >
        <PlusIcon className="size-4" />
        Add field
      </Button>
      <div className="flex justify-end gap-2 mt-2">
        <Button
          type="button"
          variant="outline"
          onClick={onCancel}
          disabled={saving}
        >
          Cancel
        </Button>
        <Button type="button" onClick={handleSave} disabled={saving}>
          {saving ? "Saving..." : "Save"}
        </Button>
      </div>
    </div>
  );
}

export default TransactionLabelEditor;
