import { createClient } from "tinacms/dist/client";
import { queries } from "./types";
export const client = createClient({ url: 'http://localhost:4001/graphql', token: 'd8f47991122c2a6eac055eb518cc9cb46f0fe1cf', queries,  });
export default client;
  