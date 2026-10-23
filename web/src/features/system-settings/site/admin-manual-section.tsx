/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { zodResolver } from "@hookform/resolvers/zod";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { z } from "zod";

import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormLabel,
} from "@/components/ui/form";
import { Switch } from "@/components/ui/switch";

import {
  SettingsForm,
  SettingsSwitchContent,
  SettingsSwitchItem,
} from "../components/settings-form-layout";
import { SettingsPageFormActions } from "../components/settings-page-context";
import { SettingsSection } from "../components/settings-section";
import { useUpdateOption } from "../hooks/use-update-option";

const adminManualSchema = z.object({ enabled: z.boolean() });

export function AdminManualSection(props: { defaultEnabled: boolean }) {
  const { t } = useTranslation();
  const update = useUpdateOption();
  const form = useForm<z.infer<typeof adminManualSchema>>({
    resolver: zodResolver(adminManualSchema),
    defaultValues: { enabled: props.defaultEnabled },
  });
  useEffect(
    () => form.reset({ enabled: props.defaultEnabled }),
    [props.defaultEnabled, form],
  );

  const save = (values: z.infer<typeof adminManualSchema>) => {
    if (values.enabled === props.defaultEnabled) return;
    update.mutate({
      key: "AdminManualEnabled",
      value: String(values.enabled),
    });
  };

  return (
    <SettingsSection title={t("Platform integration guide")}>
      <Form {...form}>
        <SettingsForm onSubmit={form.handleSubmit(save)}>
          <SettingsPageFormActions
            onSave={form.handleSubmit(save)}
            isSaving={update.isPending}
            isSaveDisabled={!form.formState.isDirty}
          />
          <FormField
            control={form.control}
            name="enabled"
            render={({ field }) => (
              <SettingsSwitchItem>
                <SettingsSwitchContent>
                  <FormLabel>
                    {t("Enable platform integration guide")}
                  </FormLabel>
                  <FormDescription>
                    {t(
                      "Allow administrators to read the platform integration guide. When disabled, root cannot preview it either.",
                    )}
                  </FormDescription>
                </SettingsSwitchContent>
                <FormControl>
                  <Switch
                    checked={field.value}
                    onCheckedChange={field.onChange}
                    disabled={update.isPending}
                  />
                </FormControl>
              </SettingsSwitchItem>
            )}
          />
        </SettingsForm>
      </Form>
    </SettingsSection>
  );
}
