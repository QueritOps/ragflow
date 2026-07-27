import { useEffect } from 'react';
import { UseFormReturn, useWatch } from 'react-hook-form';
import useGraphStore from '../../store';
import { serializeQueritFormValues } from './utils';

export function useWatchFormChange(id?: string, form?: UseFormReturn<any>) {
  const values = useWatch({ control: form?.control });
  const updateNodeForm = useGraphStore((state) => state.updateNodeForm);

  useEffect(() => {
    if (id && form?.formState.isValid) {
      updateNodeForm(id, serializeQueritFormValues(form.getValues()));
    }
  }, [form, form?.formState.isValid, id, updateNodeForm, values]);
}
