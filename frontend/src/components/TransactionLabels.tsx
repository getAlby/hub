import { PencilIcon, PlusIcon, TagIcon, XIcon } from "lucide-react";
import React from "react";
import { toast } from "sonner";
import { Badge } from "src/components/ui/badge";
import { Button } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { request } from "src/utils/request";
import { useSWRConfig } from "swr";

type Props = {
  id: number;
  labels?: Record<string, string>;
  transactionListKey: string;
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

function TransactionLabels({
  id,
  labels: initialLabels,
  transactionListKey,
}: Props) {
  const { mutate } = useSWRConfig();
  const [isEditing, setIsEditing] = React.useState(false);
  const [loading, setLoading] = React.useState(false);
  const [labels, setLabels] = React.useState<Record<string, string>>(
    initialLabels ?? {}
  );
  const [rows, setRows] = React.useState<Row[]>(() => toRows(initialLabels));

  React.useEffect(() => {
    setLabels(initialLabels ?? {});
    if (!isEditing) {
      setRows(toRows(initialLabels));
    }
  }, [initialLabels, isEditing]);

  const labelEntries = Object.entries(labels);

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

  const startEditing = () => {
    setRows(toRows(labels));
    setIsEditing(true);
  };

  const cancelEditing = () => {
    setRows(toRows(labels));
    setIsEditing(false);
  };

  const saveLabels = async () => {
    const nextLabels = rowsToLabels(rows);
    setLoading(true);

    try {
      await request(`/api/transactions/${id}/labels`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ labels: nextLabels }),
      });

      setLabels(nextLabels);
      setRows(toRows(nextLabels));
      setIsEditing(false);

      await mutate(transactionListKey);

      toast("Labels saved");
    } catch (error) {
      console.error(error);
      toast.error("Failed to save labels", {
        description: String(error),
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <div className="flex items-center justify-between">
        <p>Labels</p>
        {!isEditing && (
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={startEditing}
          >
            {labelEntries.length > 0 ? (
              <>
                <PencilIcon className="size-3" />
                Edit
              </>
            ) : (
              <>
                <TagIcon className="size-3" />
                Add labels
              </>
            )}
          </Button>
        )}
      </div>

      {isEditing ? (
        <div className="mt-2 flex flex-col gap-3">
          {rows.map((row, index) => (
            <div key={index} className="flex items-center gap-2">
              <Input
                placeholder="key"
                value={row.key}
                onChange={(e) => updateRow(index, { key: e.target.value })}
                className="w-1/3"
              />
              <Input
                placeholder="value"
                value={row.value}
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
          <div className="flex justify-end gap-2">
            <Button
              type="button"
              variant="outline"
              onClick={cancelEditing}
              disabled={loading}
            >
              Cancel
            </Button>
            <Button type="button" onClick={saveLabels} disabled={loading}>
              Save
            </Button>
          </div>
        </div>
      ) : labelEntries.length > 0 ? (
        <div className="mt-2 flex items-center gap-2 flex-wrap">
          {labelEntries.map(([key, value]) => (
            <Badge key={key} variant="secondary" className="font-normal">
              <span className="text-muted-foreground">{key}:</span>
              <span>{value}</span>
            </Badge>
          ))}
        </div>
      ) : (
        <p className="mt-1 text-sm text-muted-foreground">No labels yet.</p>
      )}
    </div>
  );
}

export default TransactionLabels;
