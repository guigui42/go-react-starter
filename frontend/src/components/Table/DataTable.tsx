import { Table, ScrollArea } from '@mantine/core'
import { type CSSProperties, type ReactNode } from 'react'
import classes from './DataTable.module.css'

export interface Column<T> {
  key: keyof T
  label: string
  render?: (value: T[keyof T], row: T) => ReactNode
}

export interface DataTableProps<T> {
  data: T[]
  columns: Column<T>[]
  getRowKey: (row: T) => string | number
  /** When true, the table header sticks to the top of the scroll area */
  stickyHeader?: boolean
  /** Optional inline style applied to each row */
  getRowStyle?: (row: T) => CSSProperties | undefined
}

const stickyTheadStyle: CSSProperties = {
  position: 'sticky',
  top: 0,
  zIndex: 1,
  backgroundColor: 'var(--mantine-color-body)',
}

export function DataTable<T>({ data, columns, getRowKey, stickyHeader = false, getRowStyle }: DataTableProps<T>) {
  return (
    <ScrollArea>
      <Table striped highlightOnHover>
        <Table.Thead style={stickyHeader ? stickyTheadStyle : undefined}>
          <Table.Tr>
            {columns.map((col) => (
              <Table.Th key={String(col.key)}>{col.label}</Table.Th>
            ))}
          </Table.Tr>
        </Table.Thead>
        <Table.Tbody>
          {data.map((row) => (
            <Table.Tr
              key={getRowKey(row)}
              className={classes.Go React StarterTableRow}
              style={getRowStyle?.(row)}
            >
              {columns.map((col) => (
                <Table.Td key={String(col.key)}>
                  {col.render ? col.render(row[col.key], row) : String(row[col.key])}
                </Table.Td>
              ))}
            </Table.Tr>
          ))}
        </Table.Tbody>
      </Table>
    </ScrollArea>
  )
}
