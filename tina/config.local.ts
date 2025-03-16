import { defineConfig } from "tinacms";
import { schema } from "./schema";

export default defineConfig({
  branch: process.env.TINA_BRANCH || "",
  clientId: process.env.TINA_PUBLIC_CLIENT_ID || "",
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
  schema,
  cmsCallback: (cms) => {
    cms.flags.set("loadCustomStore", true);
    
    cms.events.subscribe("forms:submit:success", () => {
      if (window.location.hostname === "localhost") {
        fetch("/api/rebuild", { method: "POST" });
      }
    });

    return cms;
  },
});