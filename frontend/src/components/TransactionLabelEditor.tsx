import { PlusIcon, XIcon } from "lucide-react";
import React from "react";
import { Button } from "src/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "src/components/ui/dialog";
import { Input } from "src/components/ui/input";
import { useLabelKeySuggestions } from "src/hooks/useLabelKeySuggestions";

const KEY_MAX_LENGTH = 64;
const VALUE_MAX_LENGTH = 1000;
const KEY_SUGGESTIONS_LIST_ID = "transaction-label-key-suggestions";

type Props = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  initialLabels: Record<string, string> | undefined;
  saving?: boolean;
  onSave: (labels: Record<string, string>) => void | Promise<void>;
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
  open,
  onOpenChange,
  initialLabels,
  saving,
  onSave,
}: Props) {
  const [rows, setRows] = React.useState<Row[]>(() => toRows(initialLabels));
  const keySuggestions = useLabelKeySuggestions();

  React.useEffect(() => {
    if (open) {
      setRows(toRows(initialLabels));
    }
  }, [open, initialLabels]);

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
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Edit Labels</DialogTitle>
          <DialogDescription>
            Add custom fields to label this transaction. Useful for bookkeeping
            or finding payments later.
          </DialogDescription>
        </DialogHeader>
        {keySuggestions.length > 0 && (
          <datalist id={KEY_SUGGESTIONS_LIST_ID}>
            {keySuggestions.map((key) => (
              <option key={key} value={key} />
            ))}
          </datalist>
        )}
        <div className="flex flex-col gap-3">
          {rows.map((row, index) => (
            <div key={index} className="flex items-center gap-2">
              <Input
                placeholder="key"
                value={row.key}
                maxLength={KEY_MAX_LENGTH}
                list={
                  keySuggestions.length > 0
                    ? KEY_SUGGESTIONS_LIST_ID
                    : undefined
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
        </div>
        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={saving}
          >
            Cancel
          </Button>
          <Button type="button" onClick={handleSave} disabled={saving}>
            {saving ? "Saving..." : "Save"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

export default TransactionLabelEditor;
