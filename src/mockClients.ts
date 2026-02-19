// src/mockClients.ts
//
// Mock Google API clients for testing. Activated by MOCK_AUTH=1.

const MOCK_DOC = {
  documentId: 'mock-doc-id-123',
  title: 'Mock Document',
  body: {
    content: [
      {
        paragraph: {
          elements: [
            {
              textRun: {
                content: 'Hello from the mock document.\n',
              },
            },
          ],
        },
      },
    ],
  },
  tabs: [],
};

const MOCK_SPREADSHEET_VALUES = {
  range: 'Sheet1!A1:B3',
  majorDimension: 'ROWS',
  values: [
    ['Name', 'Score'],
    ['Alice', '95'],
    ['Bob', '87'],
  ],
};

const MOCK_FILE_LIST = {
  files: [
    {
      id: 'mock-doc-id-123',
      name: 'Mock Document',
      mimeType: 'application/vnd.google-apps.document',
      modifiedTime: '2025-01-15T10:30:00.000Z',
      createdTime: '2025-01-01T08:00:00.000Z',
      webViewLink: 'https://docs.google.com/document/d/mock-doc-id-123/edit',
      owners: [{ displayName: 'Test User', emailAddress: 'test@example.com' }],
    },
    {
      id: 'mock-sheet-id-456',
      name: 'Mock Spreadsheet',
      mimeType: 'application/vnd.google-apps.spreadsheet',
      modifiedTime: '2025-01-14T09:00:00.000Z',
      createdTime: '2025-01-02T08:00:00.000Z',
      webViewLink: 'https://docs.google.com/spreadsheets/d/mock-sheet-id-456/edit',
      owners: [{ displayName: 'Test User', emailAddress: 'test@example.com' }],
    },
  ],
};

function ok(data: any) {
  return { data, status: 200, statusText: 'OK', headers: {}, config: {} };
}

export function createMockDocsClient(): any {
  return {
    documents: {
      get: async () => ok(MOCK_DOC),
      batchUpdate: async () => ok({ replies: [] }),
      create: async () => ok(MOCK_DOC),
    },
  };
}

export function createMockDriveClient(): any {
  return {
    files: {
      list: async () => ok(MOCK_FILE_LIST),
      get: async () => ok(MOCK_FILE_LIST.files[0]),
      create: async () => ok(MOCK_FILE_LIST.files[0]),
      update: async () => ok(MOCK_FILE_LIST.files[0]),
      copy: async () => ok({ ...MOCK_FILE_LIST.files[0], id: 'mock-copy-id' }),
      delete: async () => ok({}),
    },
    permissions: {
      create: async () => ok({ id: 'mock-permission-id' }),
    },
    comments: {
      list: async () => ok({ comments: [] }),
      get: async () => ok({ id: 'mock-comment-id', content: 'Mock comment' }),
      create: async () => ok({ id: 'mock-comment-id', content: 'Mock comment' }),
      delete: async () => ok({}),
    },
    replies: {
      create: async () => ok({ id: 'mock-reply-id', content: 'Mock reply' }),
    },
  };
}

export function createMockSheetsClient(): any {
  return {
    spreadsheets: {
      get: async () =>
        ok({
          spreadsheetId: 'mock-sheet-id-456',
          properties: { title: 'Mock Spreadsheet' },
          sheets: [{ properties: { sheetId: 0, title: 'Sheet1', index: 0 } }],
        }),
      values: {
        get: async () => ok(MOCK_SPREADSHEET_VALUES),
        update: async () => ok({ updatedCells: 6, updatedRows: 3 }),
        append: async () => ok({ updates: { updatedCells: 2, updatedRows: 1 } }),
        clear: async () => ok({ clearedRange: 'Sheet1!A1:B3' }),
      },
      batchUpdate: async () => ok({ replies: [] }),
      create: async () =>
        ok({
          spreadsheetId: 'mock-new-sheet-id',
          properties: { title: 'New Spreadsheet' },
        }),
    },
  };
}
