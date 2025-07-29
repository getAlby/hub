import { useMemo } from "react";
import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from "src/components/ui/pagination";
import { cn, generatePageNumbers } from "src/lib/utils";

type CustomPaginationProps = {
  page: number;
  limit: number;
  totalCount: number;
  handlePageChange(page: number): void;
};

export function CustomPagination({
  page,
  totalCount,
  limit,
  handlePageChange,
}: CustomPaginationProps) {
  const totalPages = Math.ceil(totalCount / limit);

  const pageNumbers = useMemo(() => {
    return generatePageNumbers(page, totalPages);
  }, [page, totalPages]);

  if (totalPages <= 1) {
    return null;
  }
  return (
    <div className="mt-4 self-center">
      <Pagination>
        <PaginationContent>
          <PaginationItem
            className={cn(
              page === 1 && "pointer-events-none opacity-30 dark:opacity-20"
            )}
          >
            <PaginationPrevious
              href="#"
              onClick={(e) => {
                e.preventDefault();
                handlePageChange(page - 1);
              }}
            />
          </PaginationItem>

          {pageNumbers.map((p, index) =>
            p === "ellipsis" ? (
              <PaginationItem key={index}>
                <PaginationEllipsis className="flex items-center" />
              </PaginationItem>
            ) : (
              <PaginationItem key={index}>
                <PaginationLink
                  href="#"
                  isActive={p === page}
                  onClick={(e) => {
                    e.preventDefault();
                    handlePageChange(p);
                  }}
                >
                  {p}
                </PaginationLink>
              </PaginationItem>
            )
          )}

          <PaginationItem
            className={cn(
              page === totalPages &&
                "pointer-events-none opacity-30 dark:opacity-20"
            )}
          >
            <PaginationNext
              href="#"
              onClick={(e) => {
                e.preventDefault();
                handlePageChange(page + 1);
              }}
            />
          </PaginationItem>
        </PaginationContent>
      </Pagination>
    </div>
  );
}
