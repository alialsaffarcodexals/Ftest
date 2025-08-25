package models

// Category represents a post category.
type Category struct {
    ID   int
    Name string
}

// GetCategories returns all categories.
func GetCategories() ([]Category, error) {
    rows, err := DB.Query(`SELECT id, name FROM categories`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var cs []Category
    for rows.Next() {
        var c Category
        if err := rows.Scan(&c.ID, &c.Name); err != nil {
            return nil, err
        }
        cs = append(cs, c)
    }
    return cs, nil
}
