export const roles = {
    admin: {
      name: 'Administrator',
      permissions: {
        collections: {
          blog: {
            create: true,
            read: true,
            update: true,
            delete: true,
          },
          shop: {
            create: true,
            read: true,
            update: true,
            delete: true,
          },
          settings: {
            read: true,
            update: true,
          },
        },
        media: {
          upload: true,
          delete: true,
        },
      },
    },
    editor: {
      name: 'Editor',
      permissions: {
        collections: {
          blog: {
            create: true,
            read: true,
            update: true,
            delete: false,
          },
          shop: {
            create: false,
            read: true,
            update: true,
            delete: false,
          },
          settings: {
            read: true,
            update: false,
          },
        },
        media: {
          upload: true,
          delete: false,
        },
      },
    },
    author: {
      name: 'Author',
      permissions: {
        collections: {
          blog: {
            create: true,
            read: true,
            update: 'own',
            delete: false,
          },
          shop: {
            create: false,
            read: true,
            update: false,
            delete: false,
          },
          settings: {
            read: true,
            update: false,
          },
        },
        media: {
          upload: true,
          delete: 'own',
        },
      },
    },
  };