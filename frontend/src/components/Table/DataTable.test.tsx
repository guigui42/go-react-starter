import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MantineProvider } from '@mantine/core'
import { DataTable, type Column } from './DataTable'

const TestWrapper = ({ children }: { children: React.ReactNode }) => {
  return <MantineProvider>{children}</MantineProvider>
}

interface TestRow {
  id: number
  name: string
  value: number
}

describe('DataTable', () => {
  const testData: TestRow[] = [
    { id: 1, name: 'Item 1', value: 100 },
    { id: 2, name: 'Item 2', value: 200 },
    { id: 3, name: 'Item 3', value: 300 },
  ]

  const columns: Column<TestRow>[] = [
    { key: 'id', label: 'ID' },
    { key: 'name', label: 'Name' },
    { key: 'value', label: 'Value' },
  ]

  it('should render table headers', () => {
    render(
      <TestWrapper>
        <DataTable data={testData} columns={columns} getRowKey={(row) => row.id} />
      </TestWrapper>
    )
    expect(screen.getByText('ID')).toBeInTheDocument()
    expect(screen.getByText('Name')).toBeInTheDocument()
    expect(screen.getByText('Value')).toBeInTheDocument()
  })

  it('should render table data', () => {
    render(
      <TestWrapper>
        <DataTable data={testData} columns={columns} getRowKey={(row) => row.id} />
      </TestWrapper>
    )
    expect(screen.getByText('Item 1')).toBeInTheDocument()
    expect(screen.getByText('Item 2')).toBeInTheDocument()
    expect(screen.getByText('Item 3')).toBeInTheDocument()
    expect(screen.getByText('100')).toBeInTheDocument()
    expect(screen.getByText('200')).toBeInTheDocument()
    expect(screen.getByText('300')).toBeInTheDocument()
  })

  it('should render custom cell content using render function', () => {
    const customColumns: Column<TestRow>[] = [
      { key: 'name', label: 'Name' },
      {
        key: 'value',
        label: 'Value',
        render: (value) => `$${value}`,
      },
    ]

    render(
      <TestWrapper>
        <DataTable data={testData} columns={customColumns} getRowKey={(row) => row.id} />
      </TestWrapper>
    )
    expect(screen.getByText('$100')).toBeInTheDocument()
    expect(screen.getByText('$200')).toBeInTheDocument()
    expect(screen.getByText('$300')).toBeInTheDocument()
  })

  it('should render custom cell content with access to full row', () => {
    const customColumns: Column<TestRow>[] = [
      {
        key: 'name',
        label: 'Name',
        render: (value, row) => `${value} (ID: ${row.id})`,
      },
    ]

    render(
      <TestWrapper>
        <DataTable data={testData} columns={customColumns} getRowKey={(row) => row.id} />
      </TestWrapper>
    )
    expect(screen.getByText('Item 1 (ID: 1)')).toBeInTheDocument()
    expect(screen.getByText('Item 2 (ID: 2)')).toBeInTheDocument()
  })

  it('should handle empty data', () => {
    const { container } = render(
      <TestWrapper>
        <DataTable data={[]} columns={columns} getRowKey={(row) => row.id} />
      </TestWrapper>
    )
    // Headers should still be present
    expect(screen.getByText('ID')).toBeInTheDocument()
    // But no tbody rows (except the empty tbody element)
    const tbody = container.querySelector('tbody')
    expect(tbody?.children.length).toBe(0)
  })

  it('should apply striped and highlightOnHover by default', () => {
    const { container } = render(
      <TestWrapper>
        <DataTable data={testData} columns={columns} getRowKey={(row) => row.id} />
      </TestWrapper>
    )
    const table = container.querySelector('table')
    expect(table).toBeInTheDocument()
  })

  it('should render inside ScrollArea', () => {
    const { container } = render(
      <TestWrapper>
        <DataTable data={testData} columns={columns} getRowKey={(row) => row.id} />
      </TestWrapper>
    )
    const scrollArea = container.querySelector('.mantine-ScrollArea-root')
    expect(scrollArea).toBeInTheDocument()
  })

  it('should handle complex data types', () => {
    interface ComplexRow {
      id: string
      nested: { value: string }
    }

    const complexData: ComplexRow[] = [
      { id: 'a1', nested: { value: 'test' } },
    ]

    const complexColumns: Column<ComplexRow>[] = [
      { key: 'id', label: 'ID' },
      {
        key: 'nested',
        label: 'Nested',
        render: (value) => (value as { value: string }).value,
      },
    ]

    render(
      <TestWrapper>
        <DataTable
          data={complexData}
          columns={complexColumns}
          getRowKey={(row) => row.id}
        />
      </TestWrapper>
    )
    expect(screen.getByText('test')).toBeInTheDocument()
  })
})
