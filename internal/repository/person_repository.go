package repository

import (
    "college_api/internal/model"
    "database/sql"
    "fmt"
)

// PersonRepository 人员数据访问层
type PersonRepository struct {
    db *sql.DB
}

// NewPersonRepository 创建人员仓库
func NewPersonRepository(db *sql.DB) *PersonRepository {
    return &PersonRepository{db: db}
}

// GetStaffInfo 获取政工完整信息
func (r *PersonRepository) GetStaffInfo(personID int) (*model.StaffInfo, error) {
    query := `
        SELECT 
            p.id as person_id,
            p.person_type,
            p.name,
            p.gender,
            p.mobile,
            p.email,
            p.avatar,
            p.status,
            s.staff_no,
            s.university_id,
            s.department_id,
            s.college_id,
            s.faculty_id
        FROM persons p
        LEFT JOIN staff s ON p.id = s.person_id
        WHERE p.id = ? AND p.deleted_at = 0
        LIMIT 1
    `

    var info model.StaffInfo
    var gender sql.NullInt64
    var mobile, email, avatar, staffNo sql.NullString
    var universityID sql.NullInt64
    var departmentID, collegeID, facultyID sql.NullInt64

    err := r.db.QueryRow(query, personID).Scan(
        &info.PersonID,
        &info.PersonType,
        &info.Name,
        &gender,
        &mobile,
        &email,
        &avatar,
        &info.Status,
        &staffNo,
        &universityID,
        &departmentID,
        &collegeID,
        &facultyID,
    )

    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("person not found")
        }
        return nil, err
    }

    // 处理可空字段
    if gender.Valid {
        g := int(gender.Int64)
        info.Gender = &g
    }
    if mobile.Valid {
        info.Mobile = mobile.String
    }
    if email.Valid {
        info.Email = email.String
    }
    if avatar.Valid {
        info.Avatar = avatar.String
    }
    if staffNo.Valid {
        info.StaffNo = staffNo.String
    }
    if universityID.Valid {
        info.UniversityID = int(universityID.Int64)
    }
    if departmentID.Valid {
        d := int(departmentID.Int64)
        info.DepartmentID = &d
    }
    if collegeID.Valid {
        c := int(collegeID.Int64)
        info.CollegeID = &c
    }
    if facultyID.Valid {
        f := int(facultyID.Int64)
        info.FacultyID = &f
    }

    return &info, nil
}

// GetDepartmentName 获取部门名称
func (r *PersonRepository) GetDepartmentName(deptID int) (string, error) {
    query := `SELECT department_name FROM departments WHERE id = ? AND deleted_at = 0 LIMIT 1`
    
    var name string
    err := r.db.QueryRow(query, deptID).Scan(&name)
    if err != nil {
        if err == sql.ErrNoRows {
            return "", fmt.Errorf("department not found")
        }
        return "", err
    }
    
    return name, nil
}

// GetAdminUserInfo 获取学校管理员信息
func (r *PersonRepository) GetAdminUserInfo(userID int) (*model.AdminUserInfo, error) {
    query := `
        SELECT 
            id,
            username,
            email,
            mobile,
            avatar,
            customer_id,
            status
        FROM admin_users
        WHERE id = ? AND deleted_at = 0
        LIMIT 1
    `

    var info model.AdminUserInfo
    var email, mobile, avatar sql.NullString

    err := r.db.QueryRow(query, userID).Scan(
        &info.UserID,
        &info.Username,
        &email,
        &mobile,
        &avatar,
        &info.UniversityID,
        &info.Status,
    )

    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("admin user not found")
        }
        return nil, err
    }

    // 处理可空字段
    if email.Valid {
        info.Email = email.String
    }
    if mobile.Valid {
        info.Mobile = mobile.String
    }
    if avatar.Valid {
        info.Avatar = avatar.String
    }

    return &info, nil
}
