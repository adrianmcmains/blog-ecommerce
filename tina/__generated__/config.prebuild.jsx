// tina/config.ts
import { defineConfig } from "tinacms";
var config_default = defineConfig({
  branch: process.env.TINA_BRANCH || "",
  clientId: process.env.TINA_PUBLIC_CLIENT_ID || "",
  // process.env.CLIENT_ID,
  token: process.env.TINA_TOKEN || "",
  // process.env.TOKEN,
  build: {
    outputFolder: "admin",
    publicFolder: "static"
  },
  media: {
    tina: {
      mediaRoot: "static/images",
      publicFolder: "static"
    }
  },
  // See docs on content modeling for more options
  schema: {
    collections: [
      {
        name: "blog",
        label: "Blog Posts",
        path: "content/blog",
        format: "md",
        ui: {
          filename: {
            readonly: false,
            slugify: (values) => {
              return `${values?.title?.toLowerCase().replace(/ /g, "-")}` || "";
            }
          }
        },
        fields: [
          {
            type: "string",
            label: "Title",
            name: "title",
            required: true,
            isTitle: true
          },
          {
            type: "datetime",
            name: "date",
            label: "Publication Date",
            required: true
          },
          {
            type: "image",
            name: "image",
            label: "Featured Image"
          },
          {
            type: "string",
            name: "author",
            label: "Author",
            required: true
          },
          {
            type: "string",
            name: "categories",
            label: "Categories",
            list: true,
            ui: {
              component: "tags"
            }
          },
          {
            type: "rich-text",
            name: "body",
            label: "Content",
            isBody: true,
            templates: [
              {
                name: "CodeBlock",
                label: "Code Block",
                fields: [
                  {
                    name: "language",
                    label: "Language",
                    type: "string",
                    options: ["javascript", "css", "html", "python", "go"]
                  },
                  {
                    name: "code",
                    label: "Code",
                    type: "string",
                    ui: {
                      component: "textarea"
                    }
                  }
                ]
              },
              {
                name: "BlockQuote",
                label: "Block Quote",
                fields: [
                  {
                    name: "quote",
                    label: "Quote",
                    type: "string",
                    ui: {
                      component: "textarea"
                    }
                  },
                  {
                    name: "author",
                    label: "Author",
                    type: "string"
                  }
                ]
              }
            ]
          }
        ]
      },
      {
        name: "shop",
        label: "Products",
        path: "content/shop",
        format: "md",
        ui: {
          filename: {
            readonly: false,
            slugify: (values) => {
              return `${values?.title?.toLowerCase().replace(/ /g, "-")}` || "";
            }
          }
        },
        fields: [
          {
            type: "string",
            name: "title",
            label: "Product Name",
            required: true,
            isTitle: true
          },
          {
            type: "number",
            name: "price",
            label: "Price",
            required: true
          },
          {
            type: "number",
            name: "stockQuantity",
            label: "Stock Quantity",
            required: true
          },
          {
            type: "image",
            name: "image",
            label: "Product Image",
            required: true
          },
          {
            type: "string",
            name: "categories",
            label: "Categories",
            list: true,
            ui: {
              component: "tags"
            }
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
                label: "Variant Name"
              },
              {
                type: "number",
                name: "price",
                label: "Price"
              },
              {
                type: "number",
                name: "stock",
                label: "Stock"
              }
            ]
          },
          {
            type: "rich-text",
            name: "body",
            label: "Product Description",
            isBody: true
          }
        ]
      },
      {
        name: "settings",
        label: "Site Settings",
        path: "content/settings",
        format: "json",
        ui: {
          allowedActions: {
            create: false,
            delete: false
          }
        },
        fields: [
          {
            type: "object",
            name: "site",
            label: "Site Settings",
            fields: [
              {
                type: "string",
                name: "title",
                label: "Site Title"
              },
              {
                type: "string",
                name: "description",
                label: "Site Description",
                ui: {
                  component: "textarea"
                }
              },
              {
                type: "image",
                name: "logo",
                label: "Site Logo"
              }
            ]
          },
          {
            type: "object",
            name: "social",
            label: "Social Media",
            fields: [
              {
                type: "string",
                name: "twitter",
                label: "Twitter URL"
              },
              {
                type: "string",
                name: "facebook",
                label: "Facebook URL"
              },
              {
                type: "string",
                name: "instagram",
                label: "Instagram URL"
              }
            ]
          }
        ]
      }
    ]
  },
  search: {
    tina: {
      indexerToken: process.env.TINA_SEARCH_TOKEN,
      stopwordLanguages: ["eng"]
    },
    indexBatchSize: 100,
    maxSearchIndexFieldLength: 100
  }
});
export {
  config_default as default
};
