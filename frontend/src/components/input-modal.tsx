import React, { useRef } from "react"
import { Modal, Form, Input } from "antd"
import { FormComponentProps, FormCreateOption } from "antd/lib/form"

export interface InputModalProps extends FormComponentProps<{ input?: string }> {
  visible: boolean
  initialInput?: string
  onCancel: () => void
  onInput: (input: string) => void
}

export const InputModal: React.FC<InputModalProps> = ({
  visible,
  initialInput,
  onCancel,
  onInput,
  form
}) => {
  const { getFieldDecorator } = form
  const inputRef = useRef<Input | null>(null)

  if (visible) {
    // We need to perform the `focus` after this render call (hence the setTimeout)
    // because antd itself plays with fields focus which override a straight call to
    // `focus`.
    setTimeout(() => {
      if (inputRef.current) {
        inputRef.current.focus()
      }
    }, 0)
  }

  return (
    <Modal
      visible={visible}
      title="Enter Input"
      okText="Done"
      onCancel={onCancel}
      onOk={() => {
        form.validateFields((err, values) => {
          if (err) {
            return
          }

          onInput(values.input!)
        })
      }}
    >
      <Form layout="vertical">
        <Form.Item label="Input">
          {getFieldDecorator("input", {
            initialValue: initialInput,
            validateTrigger: "onBlur",
            rules: [
              {
                required: true,
                message: "Please input your new value"
              }
            ]
          })(<Input ref={inputRef} />)}
        </Form.Item>
      </Form>
    </Modal>
  )
}

export function CreateInputModalForm(options?: FormCreateOption<InputModalProps>) {
  return Form.create<InputModalProps>(options)(InputModal)
}
