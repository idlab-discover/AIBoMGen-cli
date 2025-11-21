package bom

type Component struct {
    Name    string
    Version string
}

func Format(c Component) string {
    if c.Name == "" && c.Version == "" {
        return ""
    }
    if c.Version == "" {
        return c.Name
    }
    if c.Name == "" {
        return "@" + c.Version
    }
    return c.Name + "@" + c.Version
}
