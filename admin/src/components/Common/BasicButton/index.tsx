import * as React from "react";
import Button, { ButtonProps } from "@mui/material/Button";

interface BasicButtonProps extends ButtonProps {
  text: string;
  disabled?: boolean;
  justifyContent?: string;
  width?: string;
  margin?: string | number;
  selected?: boolean;
}

export default function BasicButton({
  text,
  onClick,
  color = "primary",
  variant = "contained",
  disabled = false,
  justifyContent = "flex-start",
  width = "100%",
  margin,
  selected = false,
  ...rest
}: BasicButtonProps) {
  return (
    <Button
      onClick={onClick}
      variant={variant}
      color={selected ? "secondary" : "primary"}
      style={{ width, justifyContent, margin }}
      disabled={disabled}
      {...rest}
    >
      {text}
    </Button>
  );
}