import { defineConfig } from "tinacms";

export default defineConfig({
  branch: process.env.TINA_BRANCH || "",
  clientId: process.env.TINA_CLIENT_ID || "",
  token: process.env.TINA_TOKEN || "",

  build: {
    outputFolder: "admin",
    publicFolder: "static",
  },

  media: {
    tina: {
      mediaRoot: "static/images",
      publicFolder: "static",
    },
  },

  schema: {
    collections: [
      // Blog Posts Collection
      {
        name: "post",
        label: "Blog Posts",
        path: "content/blog",
        format: "md",
        fields: [
          {
            type: "string",
            name: "title",
            label: "Title",
            isTitle: true,
            required: true,
          },
          {
            type: "image",
            name: "image",
            label: "Featured Image",
          },
          {
            type: "string",
            name: "author",
            label: "Author",
            required: true,
          },
          {
            type: "datetime",
            name: "date",
            label: "Publication Date",
            required: true,
          },
          {
            type: "string",
            name: "description",
            label: "Description",
            ui: {
              component: "textarea",
            },
          },
          {
            type: "string",
            name: "categories",
            label: "Categories",
            list: true,
            ui: {
              component: "tags",
            },
          },
          {
            type: "string",
            name: "tags",
            label: "Tags",
            list: true,
            ui: {
              component: "tags",
            },
          },
          {
            type: "rich-text",
            name: "body",
            label: "Content",
            isBody: true,
            templates: [
              {
                name: "callout",
                label: "Callout",
                fields: [
                  {
                    name: "type",
                    label: "Type",
                    type: "string",
                    options: ["info", "warning", "success", "error"],
                  },
                  {
                    name: "text",
                    label: "Text",
                    type: "string",
                  },
                ],
              },
              {
                name: "CodeBlock",
                label: "Code Block",
                fields: [
                  {
                    name: "language",
                    label: "Language",
                    type: "string",
                  },
                  {
                    name: "code",
                    label: "Code",
                    type: "string",
                    ui: {
                      component: "textarea",
                    },
                  },
                ],
              },
            ],
          },
        ],
      },

      // Products Collection
      {
        name: "product",
        label: "Products",
        path: "content/shop",
        format: "md",
        fields: [
          {
            type: "string",
            name: "title",
            label: "Product Name",
            isTitle: true,
            required: true,
          },
          {
            type: "image",
            name: "image",
            label: "Product Image",
          },
          {
            type: "number",
            name: "price",
            label: "Price",
            required: true,
          },
          {
            type: "number",
            name: "stockQuantity",
            label: "Stock Quantity",
            required: true,
          },
          {
            type: "string",
            name: "description",
            label: "Description",
            ui: {
              component: "textarea",
            },
          },
          {
            type: "string",
            name: "categories",
            label: "Categories",
            list: true,
            ui: {
              component: "tags",
            },
          },
          {
            type: "object",
            name: "variants",
            label: "Product Variants",
            list: true,
            fields: [
              {
                type: "string",
                name: "name",
                label: "Variant Name",
              },
              {
                type: "number",
                name: "price",
                label: "Variant Price",
              },
              {
                type: "string",
                name: "sku",
                label: "SKU",
              },
            ],
          },
          {
            type: "rich-text",
            name: "body",
            label: "Detailed Description",
            isBody: true,
          },
        ],
      },

      // Settings Collection
      {
        name: "settings",
        label: "Site Settings",
        path: "content/settings",
        format: "json",
        fields: [
          {
            type: "object",
            name: "site",
            label: "Site Settings",
            fields: [
              {
                type: "string",
                name: "title",
                label: "Site Title",
              },
              {
                type: "string",
                name: "description",
                label: "Site Description",
                ui: {
                  component: "textarea",
                },
              },
              {
                type: "image",
                name: "logo",
                label: "Site Logo",
              },
            ],
          },
          {
            type: "object",
            name: "social",
            label: "Social Media",
            fields: [
              {
                type: "string",
                name: "twitter",
                label: "Twitter URL",
              },
              {
                type: "string",
                name: "facebook",
                label: "Facebook URL",
              },
              {
                type: "string",
                name: "instagram",
                label: "Instagram URL",
              },
            ],
          },
        ],
      },
    ],
  },
});