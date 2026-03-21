import { SWRConfig } from "swr";
import { AdminDataProvider } from "../../modules/admin/model/useAdminData";

export default function AppProviders({ children }) {
  return (
    <SWRConfig
      value={{
        revalidateOnFocus: false,
        shouldRetryOnError: false,
      }}
    >
      <AdminDataProvider>{children}</AdminDataProvider>
    </SWRConfig>
  );
}
