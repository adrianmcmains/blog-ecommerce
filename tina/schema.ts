import { defineSchema, TinaTemplate } from "tinacms";

// Templates for rich text editor
const contentBlocks: TinaTemplate[] = [
  {
    name: "hero",
    label: "Hero",
    fields: [
      { name: "heading", label: "Heading", type: "string" },
      { name: "subtext", label: "Sub Text", type: "string" },
      { name: "image", label: "Image", type: "image" },
    ],
  },
  {
    name: "codeBlock",
    label: "Code Block",
    fields: [
      {
        name: "language",
        label: "Language",
        type: "string",
        options: ["javascript", "css", "html", "python", "go"],
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
];

export const schema = defineSchema({
  collections: [
    {
      label: "Blog Posts",
      name: "post",
      path: "content/blog",
      format: "md",
      ui: {
        filename: {
          readonly: false,
          slugify: values => {
            return `${values?.title?.toLowerCase().replace(/ /g, '-')}` || ""
          },
        },
      },
      fields: [
        {
          type: "string",
          label: "Title",
          name: "title",
          required: true,
          isTitle: true,
        },
        {
          type: "datetime",
          name: "date",
          label: "Publication Date",
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
          type: "string",
          name: "categories",
          label: "Categories",
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
          templates: contentBlocks,
        },
      ],
    },
    {
      label: "Products",
      name: "product",
      path: "content/shop",
      format: "md",
      ui: {
        filename: {
          readonly: false,
          slugify: values => {
            return `${values?.title?.toLowerCase().replace(/ /g, '-')}` || ""
          },
        },
      },
      fields: [
        {
          type: "string",
          name: "title",
          label: "Product Name",
          required: true,
          isTitle: true,
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
          type: "image",
          name: "image",
          label: "Product Image",
          required: true,
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
              label: "Price",
            },
            {
              type: "number",
              name: "stock",
              label: "Stock",
            },
          ],
        },
        {
          type: "rich-text",
          name: "body",
          label: "Product Description",
          isBody: true,
        },
      ],
    },
  ],
});